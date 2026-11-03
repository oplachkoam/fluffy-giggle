package entity

import (
	"errors"
	"slices"
	"strings"

	"fluffy-giggle/core/internal/domain/vo"
)

// Case is one fixture input with the values planted in it, the canaries.
type Case struct {
	id        string
	title     string
	text      string
	image     []byte
	imageMIME string
	canaries  []string
}

func NewCase(id, title, text string, image []byte, imageMIME string, canaries []string) (*Case, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("case id is empty")
	}
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("case text is empty")
	}
	if len(image) > 0 && !vo.IsImageMIME(imageMIME) {
		return nil, errors.New("case image type is not supported")
	}
	return &Case{
		id:        id,
		title:     title,
		text:      text,
		image:     slices.Clone(image),
		imageMIME: imageMIME,
		canaries:  slices.Clone(canaries),
	}, nil
}

func (c *Case) ID() string         { return c.id }
func (c *Case) Title() string      { return c.title }
func (c *Case) Text() string       { return c.text }
func (c *Case) Image() []byte      { return slices.Clone(c.image) }
func (c *Case) ImageMIME() string  { return c.imageMIME }
func (c *Case) HasImage() bool     { return len(c.image) > 0 }
func (c *Case) Canaries() []string { return slices.Clone(c.canaries) }
