package me

import (
	"net/http"

	httprender "fluffy-giggle/core/internal/platform/http/render"
)

type Handler struct{}

func New() *Handler { return &Handler{} }

type Response struct {
	Login string `json:"login"`
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	login, _, _ := r.BasicAuth()
	httprender.JSON(w, http.StatusOK, Response{Login: login})
}
