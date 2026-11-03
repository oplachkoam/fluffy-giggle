package startbenchmark

import (
	"errors"
	"net/http"

	"fluffy-giggle/core/internal/domain/errs"
	httprender "fluffy-giggle/core/internal/platform/http/render"
	"fluffy-giggle/core/internal/presentation/http/dto"
	"fluffy-giggle/core/internal/usecase/startbenchmark"
)

type Handler struct {
	uc *startbenchmark.Usecase
}

func New(uc *startbenchmark.Usecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	out, err := h.uc.Execute(r.Context())
	switch {
	case errors.Is(err, errs.ErrBenchmarkRunning):
		httprender.Conflict(w, err)
	case err != nil:
		httprender.Internal(w, r, err)
	default:
		httprender.JSON(w, http.StatusCreated, dto.FromBenchmark(out.Benchmark, nil, nil))
	}
}
