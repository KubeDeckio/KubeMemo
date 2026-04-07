package service

import (
	"testing"
	"time"

	"github.com/KubeDeckio/KubeMemo/internal/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDuplicateActivityInWindowMatchesIdenticalChange(t *testing.T) {
	now := time.Now().UTC()
	notes := []model.Note{
		{
			ID:        "activity-1",
			CreatedAt: ptrTime(now.Add(-30 * time.Second)),
			Activity: map[string]any{
				"action":    "scale",
				"fieldPath": "spec.replicas",
				"oldValue":  "2",
				"newValue":  "3",
			},
		},
	}

	duplicate := duplicateActivityInWindow(notes, ActivityEvent{
		Action:     "scale",
		FieldPath:  "spec.replicas",
		OldValue:   "2",
		NewValue:   "3",
		DetectedAt: now,
	}, time.Minute)

	if !duplicate {
		t.Fatalf("expected identical activity inside dedupe window to be treated as duplicate")
	}
}

func TestDuplicateActivityInWindowIgnoresExpiredWindow(t *testing.T) {
	now := time.Now().UTC()
	notes := []model.Note{
		{
			ID:        "activity-1",
			CreatedAt: ptrTime(now.Add(-2 * time.Minute)),
			Activity: map[string]any{
				"action":    "scale",
				"fieldPath": "spec.replicas",
				"oldValue":  "2",
				"newValue":  "3",
			},
		},
	}

	duplicate := duplicateActivityInWindow(notes, ActivityEvent{
		Action:     "scale",
		FieldPath:  "spec.replicas",
		OldValue:   "2",
		NewValue:   "3",
		DetectedAt: now,
	}, time.Minute)

	if duplicate {
		t.Fatalf("expected identical activity outside dedupe window to be captured again")
	}
}

func TestDuplicateActivityInWindowIgnoresDifferentChange(t *testing.T) {
	now := time.Now().UTC()
	notes := []model.Note{
		{
			ID:        "activity-1",
			CreatedAt: ptrTime(now.Add(-30 * time.Second)),
			Activity: map[string]any{
				"action":    "scale",
				"fieldPath": "spec.replicas",
				"oldValue":  "2",
				"newValue":  "3",
			},
		},
	}

	duplicate := duplicateActivityInWindow(notes, ActivityEvent{
		Action:     "scale",
		FieldPath:  "spec.replicas",
		OldValue:   "3",
		NewValue:   "4",
		DetectedAt: now,
	}, time.Minute)

	if duplicate {
		t.Fatalf("expected different change values not to be deduplicated")
	}
}

func TestIsExpiredStatus(t *testing.T) {
	if !isExpiredStatus(&metav1.Status{Reason: metav1.StatusReasonExpired}) {
		t.Fatalf("expected expired reason to be treated as expired status")
	}
	if !isExpiredStatus(&metav1.Status{Code: 410}) {
		t.Fatalf("expected 410 status code to be treated as expired status")
	}
	if isExpiredStatus(&metav1.Status{Code: 500}) {
		t.Fatalf("did not expect non-expired status to match")
	}
}

func TestNextWatchBackoffCaps(t *testing.T) {
	delay := time.Second
	for range 10 {
		delay = nextWatchBackoff(delay)
	}
	if delay != 15*time.Second {
		t.Fatalf("expected backoff to cap at 15s, got %s", delay)
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
