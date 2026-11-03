package streamevents

import (
	"context"

	"fluffy-giggle/core/internal/domain/port"
)

type Usecase struct {
	events port.EventBus
}

func New(events port.EventBus) *Usecase {
	return &Usecase{events: events}
}

func (u *Usecase) Execute(ctx context.Context) <-chan port.Event {
	return u.events.Subscribe(ctx)
}
