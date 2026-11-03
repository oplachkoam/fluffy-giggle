package resolvereview

import (
	"encoding/json"
	"errors"
	"net/http"
	"uuid"

	"github.com/go-chi/chi/v5"

	"fluffy-giggle/core/internal/domain/errs"
	httprender "fluffy-giggle/core/internal/platform/http/render"
	"fluffy-giggle/core/internal/presentation/http/dto"
	"fluffy-giggle/core/internal/usecase/resolvereview"
)

type Handler struct {
	uc *resolvereview.Usecase
}

func New(uc *resolvereview.Usecase) *Handler {
	return &Handler{uc: uc}
}

type Request struct {
	Decision string `json:"decision"`
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httprender.NotFound(w, errs.ErrRunNotFound)
		return
	}
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httprender.InvalidJSON(w)
		return
	}
	out, err := h.uc.Execute(r.Context(), resolvereview.Input{RunID: id, Decision: req.Decision})
	switch {
	case errors.Is(err, errs.ErrValidationError):
		httprender.ValidationError(w, err, req)
	case errors.Is(err, errs.ErrRunNotFound):
		httprender.NotFound(w, err)
	case errors.Is(err, errs.ErrRunNotAwaitingReview):
		httprender.Conflict(w, err)
	case err != nil:
		httprender.Internal(w, r, err)
	default:
		httprender.JSON(w, http.StatusOK, dto.FromRun(out.Run, nil))
	}
}
