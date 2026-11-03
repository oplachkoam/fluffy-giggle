package getrunimage

import (
	"errors"
	"net/http"
	"uuid"

	"github.com/go-chi/chi/v5"

	"fluffy-giggle/core/internal/domain/errs"
	httprender "fluffy-giggle/core/internal/platform/http/render"
	"fluffy-giggle/core/internal/usecase/getrunimage"
)

type Handler struct {
	uc *getrunimage.Usecase
}

func New(uc *getrunimage.Usecase) *Handler {
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
	case errors.Is(err, errs.ErrRunNotFound), errors.Is(err, errs.ErrRunHasNoImage):
		httprender.NotFound(w, err)
	case err != nil:
		httprender.Internal(w, r, err)
	default:
		w.Header().Set("Content-Type", out.MIME)
		w.Header().Set("Cache-Control", "private, max-age=3600")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(out.Data)
	}
}
