package listcases

import (
	"net/http"

	httprender "fluffy-giggle/core/internal/platform/http/render"
	"fluffy-giggle/core/internal/presentation/http/dto"
	"fluffy-giggle/core/internal/usecase/listcases"
)

type Handler struct {
	uc *listcases.Usecase
}

func New(uc *listcases.Usecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Handle(w http.ResponseWriter, _ *http.Request) {
	cases := h.uc.Execute()
	out := make([]dto.Case, 0, len(cases))
	for _, c := range cases {
		out = append(out, dto.FromCase(c))
	}
	httprender.JSON(w, http.StatusOK, out)
}
