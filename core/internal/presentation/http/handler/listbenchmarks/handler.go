package listbenchmarks

import (
	"net/http"

	httprender "fluffy-giggle/core/internal/platform/http/render"
	"fluffy-giggle/core/internal/presentation/http/dto"
	"fluffy-giggle/core/internal/usecase/listbenchmarks"
)

type Handler struct {
	uc *listbenchmarks.Usecase
}

func New(uc *listbenchmarks.Usecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	benchmarks, err := h.uc.Execute(r.Context())
	if err != nil {
		httprender.Internal(w, r, err)
		return
	}
	out := make([]dto.Benchmark, 0, len(benchmarks))
	for _, b := range benchmarks {
		out = append(out, dto.FromBenchmark(b, nil, nil))
	}
	httprender.JSON(w, http.StatusOK, out)
}
