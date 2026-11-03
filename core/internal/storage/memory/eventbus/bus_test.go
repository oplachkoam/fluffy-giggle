package eventbus

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"fluffy-giggle/core/internal/domain/port"
)

func TestBus(t *testing.T) {
	t.Parallel()

	bus := New()
	ctx, cancel := context.WithCancel(t.Context())
	ch := bus.Subscribe(ctx)

	bus.Publish(port.Event{Name: "run", Data: 1})
	select {
	case event := <-ch:
		assert.Equal(t, "run", event.Name)
	case <-time.After(time.Second):
		t.Fatal("no event")
	}

	cancel()
	select {
	case _, open := <-ch:
		assert.False(t, open)
	case <-time.After(time.Second):
		t.Fatal("channel not closed")
	}
	bus.Publish(port.Event{Name: "after"})
}
