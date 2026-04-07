package service

import (
	"testing"

	"github.com/KubeDeckio/KubeMemo/internal/model"
)

func TestSlugify(t *testing.T) {
	got := slugify("Orders API warm-up behavior")
	if got != "orders-api-warm-up-behavior" {
		t.Fatalf("unexpected slug: %s", got)
	}
}

func TestAnnotationStateForNotes(t *testing.T) {
	notes := []model.Note{
		{ID: "runtime-note", StoreType: "Runtime", Title: "Runtime note", Summary: "Temporary context"},
	}

	state := annotationStateForNotes(notes, "", model.Target{
		Mode:       "resource",
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Namespace:  "prod",
		Name:       "orders-api",
	})
	if state[annotationEnabled] != "true" {
		t.Fatalf("expected %s to be true", annotationEnabled)
	}
	if state[annotationHasNote] != "true" {
		t.Fatalf("expected %s to be true", annotationHasNote)
	}
	if state[annotationRuntimeEnabled] != "true" {
		t.Fatalf("expected %s to be true", annotationRuntimeEnabled)
	}
	if state[annotationNoteCount] != "1" {
		t.Fatalf("unexpected note count: %s", state[annotationNoteCount])
	}
	if state[annotationRuntimeCount] != "1" {
		t.Fatalf("unexpected runtime count: %s", state[annotationRuntimeCount])
	}
	if state[annotationSummary] != "Temporary context" {
		t.Fatalf("unexpected summary value: %s", state[annotationSummary])
	}
	if state[annotationView] != "kubememo show --kind Deployment --namespace prod --name orders-api" {
		t.Fatalf("unexpected view command: %s", state[annotationView])
	}
}

func TestAnnotationStateOmitsSummaryForMultipleNotes(t *testing.T) {
	notes := []model.Note{
		{ID: "b-note", StoreType: "Runtime", Title: "Runtime note", Summary: "Temporary context"},
		{ID: "a-note", StoreType: "Durable", Title: "Durable note", Summary: "Primary summary"},
	}

	state := annotationStateForNotes(notes, "", model.Target{
		Mode:       "resource",
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Namespace:  "prod",
		Name:       "orders-api",
	})
	if state[annotationEnabled] != "true" {
		t.Fatalf("expected %s to be true", annotationEnabled)
	}
	if state[annotationHasNote] != "true" {
		t.Fatalf("expected %s to be true", annotationHasNote)
	}
	if state[annotationRuntimeEnabled] != "true" {
		t.Fatalf("expected %s to be true", annotationRuntimeEnabled)
	}
	if state[annotationNoteCount] != "2" {
		t.Fatalf("unexpected note count: %s", state[annotationNoteCount])
	}
	if _, ok := state[annotationSummary]; ok {
		t.Fatalf("summary should be omitted when multiple memos are attached")
	}
}

func TestAnnotationSummaryTruncatesAndPrefersDurable(t *testing.T) {
	long := "This is a very long durable summary that should be truncated before it becomes noisy in resource annotations while still remaining useful to operators."
	state := annotationStateForNotes([]model.Note{
		{ID: "runtime-note", StoreType: "Runtime", Summary: "Runtime summary"},
		{ID: "durable-note", StoreType: "Durable", Summary: long},
	}, "", model.Target{
		Mode:       "resource",
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Namespace:  "prod",
		Name:       "orders-api",
	})

	if _, ok := state[annotationSummary]; ok {
		t.Fatalf("summary should be omitted when multiple memos are attached")
	}
}

func TestAnnotationSummaryPrefersExplicitMemo(t *testing.T) {
	state := annotationStateForNotes([]model.Note{
		{ID: "older-note", StoreType: "Durable", Summary: "Older summary"},
		{ID: "preferred-note", StoreType: "Durable", Summary: "Preferred summary"},
	}, "preferred-note", model.Target{
		Mode:       "resource",
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Namespace:  "prod",
		Name:       "orders-api",
	})

	if _, ok := state[annotationSummary]; ok {
		t.Fatalf("summary should be omitted when multiple durable memos are attached")
	}
	if state[annotationNoteCount] != "2" {
		t.Fatalf("unexpected note count: %s", state[annotationNoteCount])
	}
}

func TestAnnotationSummaryTruncatesSingleMemo(t *testing.T) {
	long := "This is a very long durable summary that should be truncated before it becomes noisy in resource annotations while still remaining useful to operators."
	state := annotationStateForNotes([]model.Note{
		{ID: "durable-note", StoreType: "Durable", Summary: long},
	}, "", model.Target{
		Mode:       "resource",
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Namespace:  "prod",
		Name:       "orders-api",
	})
	if len(state[annotationSummary]) > maxAnnotationSummaryLen {
		t.Fatalf("expected truncated summary, got length %d", len(state[annotationSummary]))
	}
	if state[annotationSummary][len(state[annotationSummary])-3:] != "..." {
		t.Fatalf("expected truncated summary to end with ellipsis, got %q", state[annotationSummary])
	}
}
