// Package fixture is the synthetic set: twenty claim letters with planted values.
// Five of them are also scans, rendered from the same text, for the vision path.
package fixture

import (
	"embed"
	"fmt"
	"io/fs"

	"fluffy-giggle/core/internal/domain/entity"
	"fluffy-giggle/core/internal/domain/port"
)

//go:embed images/*.png
var images embed.FS

const ImageCases = 5

type Catalog struct {
	cases []*entity.Case
	byID  map[string]*entity.Case
}

var _ port.CaseCatalog = (*Catalog)(nil)

func New() (*Catalog, error) {
	c := &Catalog{byID: map[string]*entity.Case{}}
	for i, letter := range letters {
		n := i + 1
		v := generate(n)
		id := fmt.Sprintf("case-%02d", n)
		var image []byte
		mime := ""
		if n <= ImageCases {
			data, err := fs.ReadFile(images, "images/"+id+".png")
			if err != nil {
				return nil, fmt.Errorf("fixture image %s: %w", id, err)
			}
			image, mime = data, "image/png"
		}
		entry, err := entity.NewCase(id, letter.title, v.fill(letter.template), image, mime, v.canaries())
		if err != nil {
			return nil, fmt.Errorf("fixture %s: %w", id, err)
		}
		c.cases = append(c.cases, entry)
		c.byID[id] = entry
	}
	return c, nil
}

func (c *Catalog) List() []*entity.Case {
	return append([]*entity.Case(nil), c.cases...)
}

func (c *Catalog) Get(id string) *entity.Case {
	return c.byID[id]
}

// Texts renders every letter without images; the scans are made from these.
func Texts() map[string]string {
	out := map[string]string{}
	for i, letter := range letters {
		out[fmt.Sprintf("case-%02d", i+1)] = generate(i + 1).fill(letter.template)
	}
	return out
}
