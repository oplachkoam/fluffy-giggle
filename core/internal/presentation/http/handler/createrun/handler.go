package createrun

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/errs"
	httprender "fluffy-giggle/core/internal/platform/http/render"
	"fluffy-giggle/core/internal/presentation/http/dto"
	"fluffy-giggle/core/internal/usecase/createrun"
)

type Handler struct {
	uc *createrun.Usecase
}

func New(uc *createrun.Usecase) *Handler {
	return &Handler{uc: uc}
}

// Request is the JSON form; the multipart form carries the same fields plus an
// `image` file.
type Request struct {
	AdvisorMode string `json:"advisor_mode"`
	Text        string `json:"text"`
	CaseID      string `json:"case_id"`
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	in, ok := h.decode(w, r)
	if !ok {
		return
	}
	out, err := h.uc.Execute(r.Context(), in)
	switch {
	case errors.Is(err, errs.ErrValidationError):
		httprender.ValidationError(w, err, Request{})
	case errors.Is(err, errs.ErrCaseNotFound):
		httprender.NotFound(w, err)
	case err != nil:
		httprender.Internal(w, r, err)
	default:
		httprender.JSON(w, http.StatusCreated, dto.FromRun(out.Run, nil))
	}
}

func (h *Handler) decode(w http.ResponseWriter, r *http.Request) (createrun.Input, bool) {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(entity.MaxImageBytes + 1<<20); err != nil {
			httprender.Error(w, http.StatusBadRequest, errors.New("invalid multipart form"))
			return createrun.Input{}, false
		}
		in := createrun.Input{
			AdvisorMode: r.FormValue("advisor_mode"),
			Text:        r.FormValue("text"),
			CaseID:      r.FormValue("case_id"),
		}
		file, _, err := r.FormFile("image")
		if err == nil {
			defer func() { _ = file.Close() }()
			data, err := io.ReadAll(io.LimitReader(file, entity.MaxImageBytes+1))
			if err != nil {
				httprender.Error(w, http.StatusBadRequest, errors.New("cannot read image"))
				return createrun.Input{}, false
			}
			in.Image = data
			in.ImageMIME = http.DetectContentType(data)
		}
		return in, true
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httprender.InvalidJSON(w)
		return createrun.Input{}, false
	}
	return createrun.Input{AdvisorMode: req.AdvisorMode, Text: req.Text, CaseID: req.CaseID}, true
}
