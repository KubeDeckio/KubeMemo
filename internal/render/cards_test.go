package render

import (
	"strings"
	"testing"
	"time"

	"github.com/KubeDeckio/KubeMemo/internal/model"
)

func TestRenderNoteCardIncludesExpectedSections(t *testing.T) {
	expires := time.Now().UTC().Add(2 * time.Hour)
	note := model.Note{
		Title:     "Orders API warm-up behavior",
		StoreType: "Runtime",
		NoteType:  "activity",
		Severity:  "info",
		Kind:      "Deployment",
		Namespace: "prod",
		Name:      "orders-api",
		Summary:   "Replicas changed from 2 to 5 during investigation",
		Content:   "Temporary runtime note for testing rendering.",
		CreatedBy: "kubernetes-admin",
		ExpiresAt: &expires,
	}

	rendered := RenderNoteCard(note, 56, false)
	for _, fragment := range []string{
		"Orders API warm-up behavior",
		"ACTIVITY",
		"expires",
		"summary",
		"notes",
	} {
		if !strings.Contains(rendered, fragment) {
			t.Fatalf("expected rendered card to contain %q", fragment)
		}
	}
}

func TestTargetLabelVariants(t *testing.T) {
	cases := []struct {
		note model.Note
		want string
	}{
		{note: model.Note{TargetMode: "resource", Kind: "Deployment", Namespace: "prod", Name: "orders-api"}, want: "Deployment/prod/orders-api"},
		{note: model.Note{TargetMode: "namespace", Namespace: "payments-prod"}, want: "namespace/payments-prod"},
		{note: model.Note{TargetMode: "app", AppName: "orders-api", AppInstance: "prod"}, want: "orders-api/prod"},
	}

	for _, tc := range cases {
		got := TargetLabel(tc.note)
		if got != tc.want {
			t.Fatalf("expected %q, got %q", tc.want, got)
		}
	}
}
