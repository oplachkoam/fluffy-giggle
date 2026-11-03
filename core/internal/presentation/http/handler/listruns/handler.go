package listruns

import (
	"net/http"
	"uuid"

	httprender "fluffy-giggle/core/internal/platform/http/render"
	"fluffy-giggle/core/internal/presentation/http/dto"
	"fluffy-giggle/core/internal/usecase/listruns"
)

type Handler struct {
	uc *listruns.Usecase
}

func New(uc *listruns.Usecase) *Handler {
	return &Handler{uc: uc}
}

type Response struct {
	Runs []dto.Run `json:"runs"`
	Next string    `json:"next"`
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	before := uuid.Nil()
	if raw := r.URL.Query().Get("before"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			httprender.ValidationError(w, err, nil)
			return
		}
		before = parsed
	}
	out, err := h.uc.Execute(r.Context(), listruns.Input{Before: before})
	if err != nil {
		httprender.Internal(w, r, err)
		return
	}
	resp := Response{Runs: make([]dto.Run, 0, len(out.Runs))}
	for _, run := range out.Runs {
		resp.Runs = append(resp.Runs, dto.FromRun(run, nil))
	}
	if out.Next != uuid.Nil() {
		resp.Next = out.Next.String()
	}
	httprender.JSON(w, http.StatusOK, resp)
}
