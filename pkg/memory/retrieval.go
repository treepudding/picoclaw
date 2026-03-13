package memory

import (
	"context"

	"github.com/sipeed/picoclaw/pkg/providers"
)

// Retrieval provides three-layer memory retrieval (L1 short-term, L2 structured summary).
// L3 long-term memory (e.g. MEMORY.md, daily notes) is handled by the agent's MemoryStore
// and is not part of the session store.
type Retrieval struct {
	store Store
}

// NewRetrieval creates a Retrieval that uses the given store for L1 and L2.
func NewRetrieval(store Store) *Retrieval {
	return &Retrieval{store: store}
}

// L1L2Result holds the result of L1 (recent messages) and L2 (structured summary).
type L1L2Result struct {
	// L1Messages are the last N messages (short-term, session-level).
	L1Messages []providers.Message
	// L2Summary is the structured summary (mid-term), or nil if none.
	L2Summary *StructuredSummary
}

// GetL1L2 returns L1 (last shortTermN messages) and L2 (structured summary) for the session.
// L1 is a suffix of the full history; L2 is loaded from the structured summary file.
func (r *Retrieval) GetL1L2(ctx context.Context, sessionKey string, shortTermN int) (*L1L2Result, error) {
	history, err := r.store.GetHistory(ctx, sessionKey)
	if err != nil {
		return nil, err
	}
	l1 := LastN(history, shortTermN)
	l2, err := r.store.GetStructuredSummary(ctx, sessionKey)
	if err != nil {
		return nil, err
	}
	return &L1L2Result{L1Messages: l1, L2Summary: l2}, nil
}

// LastN returns the last n messages from history. If n <= 0 or n >= len(history),
// returns history as-is or full history respectively.
func LastN(history []providers.Message, n int) []providers.Message {
	if n <= 0 {
		return history
	}
	if len(history) <= n {
		return history
	}
	return history[len(history)-n:]
}
