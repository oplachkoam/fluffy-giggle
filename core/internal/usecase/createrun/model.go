package createrun

import "fluffy-giggle/core/internal/domain/entity"

type Input struct {
	AdvisorMode string `validate:"required,oneof=off on placebo"`
	Text        string
	Image       []byte
	ImageMIME   string
	// CaseID picks a fixture case instead of Text or Image.
	CaseID string
}

type Output struct {
	Run *entity.Run
}
