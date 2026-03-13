package memory

import "time"

// Decision records an important decision made during the conversation.
type Decision struct {
	Context string `json:"context"` // decision context
	Choice  string `json:"choice"`  // choice made
	Reason  string `json:"reason"`  // reason for the choice
}

// StructuredSummary holds a structured, queryable summary of a session
// for the three-layer memory architecture (L2 mid-term memory).
type StructuredSummary struct {
	// Facts holds key facts extracted from the conversation.
	Facts []string `json:"facts,omitempty"`
	// Decisions holds important decisions with context.
	Decisions []Decision `json:"decisions,omitempty"`
	// Preferences holds learned user preferences (e.g. "answer briefly", "no markdown").
	Preferences []string `json:"preferences,omitempty"`
	// CurrentTask describes the task currently in progress.
	CurrentTask string `json:"current_task,omitempty"`
	// PendingItems lists unfinished to-dos or follow-ups.
	PendingItems []string `json:"pending_items,omitempty"`
	// UpdatedAt is the last update time.
	UpdatedAt time.Time `json:"updated_at"`
}

// IsEmpty returns true if the summary has no meaningful content.
func (s *StructuredSummary) IsEmpty() bool {
	return len(s.Facts) == 0 &&
		len(s.Decisions) == 0 &&
		len(s.Preferences) == 0 &&
		s.CurrentTask == "" &&
		len(s.PendingItems) == 0
}
