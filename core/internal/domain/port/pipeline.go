package port

import (
	"context"
	"uuid"

	"fluffy-giggle/core/internal/domain/vo"
)

// Detector is the deterministic gate: code that finds PII by shape.
type Detector interface {
	Detect(text string) []vo.Span
}

// RunQueue hands a run to the background workers.
type RunQueue interface {
	Enqueue(id uuid.UUID)
}

// Event is what the SSE stream carries: the name and a JSON-able payload.
type Event struct {
	Name string
	Data any
}

type EventBus interface {
	Publish(event Event)
	// Subscribe returns a channel that closes when ctx ends or the bus stops.
	Subscribe(ctx context.Context) <-chan Event
}
