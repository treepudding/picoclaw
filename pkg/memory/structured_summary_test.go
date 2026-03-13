package memory

import (
	"testing"
	"time"
)

func TestStructuredSummary_IsEmpty(t *testing.T) {
	empty := &StructuredSummary{}
	if !empty.IsEmpty() {
		t.Error("expected IsEmpty true for zero value")
	}
	withFacts := &StructuredSummary{Facts: []string{"a"}}
	if withFacts.IsEmpty() {
		t.Error("expected IsEmpty false when Facts set")
	}
	withTask := &StructuredSummary{CurrentTask: "task"}
	if withTask.IsEmpty() {
		t.Error("expected IsEmpty false when CurrentTask set")
	}
}

func TestStructuredSummary_UpdatedAt(t *testing.T) {
	s := &StructuredSummary{
		Facts:     []string{"f1"},
		UpdatedAt: time.Now(),
	}
	if s.IsEmpty() {
		t.Error("expected non-empty")
	}
}
