package getrun

import (
	"errors"
	"net/http"
	"uuid"

	"github.com/go-chi/chi/v5"

	"fluffy-giggle/core/internal/domain/errs"
	httprender "fluffy-giggle/core/internal/platform/http/render"
	"fluffy-giggle/core/internal/presentation/http/dto"
	"fluffy-giggle/core/internal/usecase/getrun"
)

type Handler struct {
	uc *getrun.Usecase
}

func New(uc *getrun.Usecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httprender.NotFound(w, errs.ErrRunNotFound)
		return
	}
	out, err := h.uc.Execute(r.Context(), id)
	switch {
	case errors.Is(err, errs.ErrRunNotFound):
		httprender.NotFound(w, err)
	case err != nil:
		httprender.Internal(w, r, err)
	default:
		httprender.JSON(w, http.StatusOK, dto.FromRun(out.Run, out.Steps))
	}
}
