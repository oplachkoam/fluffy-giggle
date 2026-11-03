package runpipeline

import (
	"context"
	"log/slog"
	"sync"
	"uuid"

	"fluffy-giggle/core/internal/domain/port"
)

const queueSize = 1024

// Worker runs queued runs on a fixed number of goroutines; bootstrap starts and
// stops it with the application.
type Worker struct {
	uc          *Usecase
	concurrency int
	queue       chan uuid.UUID
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

var _ port.RunQueue = (*Worker)(nil)

func NewWorker(uc *Usecase, concurrency int) *Worker {
	return &Worker{uc: uc, concurrency: concurrency, queue: make(chan uuid.UUID, queueSize)}
}

func (w *Worker) Enqueue(id uuid.UUID) {
	select {
	case w.queue <- id:
	default:
		slog.Error("run queue is full, run dropped", slog.String("run_id", id.String()))
	}
}

func (w *Worker) Start(context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	for range w.concurrency {
		w.wg.Add(1)
		go w.loop(ctx)
	}
	return nil
}

func (w *Worker) Stop(ctx context.Context) error {
	if w.cancel != nil {
		w.cancel()
	}
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
	}
	return nil
}

func (w *Worker) loop(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case id := <-w.queue:
			if err := w.uc.Execute(ctx, id); err != nil {
				slog.ErrorContext(ctx, "run execution failed",
					slog.String("run_id", id.String()), slog.String("error", err.Error()))
			}
		}
	}
}
