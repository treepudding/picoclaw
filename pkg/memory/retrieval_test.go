package memory

import (
	"context"
	"testing"

	"github.com/sipeed/picoclaw/pkg/providers"
)

func TestLastN(t *testing.T) {
	history := []providers.Message{
		{Role: "user", Content: "1"},
		{Role: "assistant", Content: "2"},
		{Role: "user", Content: "3"},
		{Role: "assistant", Content: "4"},
		{Role: "user", Content: "5"},
	}
	tests := []struct {
		name     string
		n        int
		wantLen  int
		wantLast string
	}{
		{"n=0 returns all", 0, 5, "5"},
		{"n=2 returns last 2", 2, 2, "5"},
		{"n=5 returns all", 5, 5, "5"},
		{"n=10 returns all", 10, 5, "5"},
		{"n negative returns all", -1, 5, "5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LastN(history, tt.n)
			if len(got) != tt.wantLen {
				t.Errorf("LastN(_, %d) len = %d, want %d", tt.n, len(got), tt.wantLen)
			}
			if len(got) > 0 && got[len(got)-1].Content != tt.wantLast {
				t.Errorf("last content = %q, want %q", got[len(got)-1].Content, tt.wantLast)
			}
		})
	}
}

func TestRetrieval_GetL1L2(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_ = store.AddMessage(ctx, "s1", "user", "hello")
	_ = store.AddMessage(ctx, "s1", "assistant", "hi")
	r := NewRetrieval(store)
	result, err := r.GetL1L2(ctx, "s1", 10)
	if err != nil {
		t.Fatalf("GetL1L2: %v", err)
	}
	if len(result.L1Messages) != 2 {
		t.Errorf("L1 messages = %d, want 2", len(result.L1Messages))
	}
	if result.L2Summary != nil {
		t.Errorf("L2 summary should be nil for new session, got %+v", result.L2Summary)
	}
}

func TestRetrieval_GetL1L2_WithStructuredSummary(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_ = store.AddMessage(ctx, "s1", "user", "hello")
	structured := &StructuredSummary{
		Facts:        []string{"user said hello"},
		CurrentTask:  "greeting",
		PendingItems: []string{"follow up"},
	}
	_ = store.SetStructuredSummary(ctx, "s1", structured)
	r := NewRetrieval(store)
	result, err := r.GetL1L2(ctx, "s1", 10)
	if err != nil {
		t.Fatalf("GetL1L2: %v", err)
	}
	if len(result.L1Messages) != 1 {
		t.Errorf("L1 messages = %d, want 1", len(result.L1Messages))
	}
	if result.L2Summary == nil {
		t.Fatal("L2 summary should be set")
	}
	if result.L2Summary.CurrentTask != "greeting" {
		t.Errorf("L2 CurrentTask = %q, want greeting", result.L2Summary.CurrentTask)
	}
	if len(result.L2Summary.PendingItems) != 1 || result.L2Summary.PendingItems[0] != "follow up" {
		t.Errorf("L2 PendingItems = %v", result.L2Summary.PendingItems)
	}
}
