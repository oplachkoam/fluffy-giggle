// Package events streams run, step and benchmark changes as server-sent events.
package events

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/port"
	"fluffy-giggle/core/internal/presentation/http/dto"
	"fluffy-giggle/core/internal/usecase/streamevents"
)

const keepAlive = 15 * time.Second

type Handler struct {
	uc *streamevents.Usecase
}

func New(uc *streamevents.Usecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	// The server's write timeout would cut a long stream; this connection has none.
	_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	events := h.uc.Execute(r.Context())
	ticker := time.NewTicker(keepAlive)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			_, _ = fmt.Fprint(w, ": keep-alive\n\n")
			flusher.Flush()
		case event, open := <-events:
			if !open {
				return
			}
			payload, ok := render(event)
			if !ok {
				continue
			}
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Name, payload)
			flusher.Flush()
		}
	}
}

func render(event port.Event) ([]byte, bool) {
	var body any
	switch data := event.Data.(type) {
	case *entity.Run:
		body = dto.FromRun(data, nil)
	case *entity.Step:
		body = dto.FromStep(data)
	case *entity.Benchmark:
		body = dto.FromBenchmark(data, nil, nil)
	default:
		return nil, false
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, false
	}
	return payload, true
}
