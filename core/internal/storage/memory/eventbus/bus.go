// Package eventbus fans pipeline events out to SSE subscribers. A slow subscriber
// loses events rather than slowing the pipeline; the page refetches on reconnect.
package eventbus

import (
	"context"
	"log/slog"
	"sync"

	"fluffy-giggle/core/internal/domain/port"
)

const subscriberBuffer = 256

type Bus struct {
	mu   sync.Mutex
	subs map[chan port.Event]struct{}
}

var _ port.EventBus = (*Bus)(nil)

func New() *Bus {
	return &Bus{subs: map[chan port.Event]struct{}{}}
}

func (b *Bus) Publish(event port.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs {
		select {
		case ch <- event:
		default:
			slog.Warn("event dropped for a slow subscriber", slog.String("event", event.Name))
		}
	}
}

func (b *Bus) Subscribe(ctx context.Context) <-chan port.Event {
	ch := make(chan port.Event, subscriberBuffer)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()

	go func() {
		<-ctx.Done()
		b.mu.Lock()
		delete(b.subs, ch)
		close(ch)
		b.mu.Unlock()
	}()

	return ch
}
