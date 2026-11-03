package fixture

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fluffy-giggle/core/internal/anonymizer"
	"fluffy-giggle/core/internal/domain/vo"
)

func TestCatalog(t *testing.T) {
	t.Parallel()

	catalog, err := New()
	require.NoError(t, err)
	cases := catalog.List()
	require.Len(t, cases, 20)

	detector := anonymizer.New()
	images := 0
	for _, c := range cases {
		assert.NotContains(t, c.Text(), "{{", c.ID())
		for _, canary := range c.Canaries() {
			assert.NotEmpty(t, canary, c.ID())
		}
		assert.Contains(t, c.Text(), c.Canaries()[0], "nominative name is in every letter")
		assert.NotEmpty(t, detector.Detect(c.Text()), "gate 1 finds something in %s", c.ID())
		if c.HasImage() {
			images++
			assert.Equal(t, "image/png", c.ImageMIME())
		}
		assert.Same(t, c, catalog.Get(c.ID()))
	}
	assert.Equal(t, ImageCases, images)
	assert.Nil(t, catalog.Get("case-99"))
}

func TestGenerateIsDeterministicAndValid(t *testing.T) {
	t.Parallel()

	assert.Equal(t, generate(3), generate(3))
	assert.NotEqual(t, generate(3).phone, generate(4).phone)

	detector := anonymizer.New()
	for n := 1; n <= 20; n++ {
		v := generate(n)
		text := strings.Join(
			[]string{"СНИЛС " + v.snils, "ИНН " + v.inn, "карта " + v.card, "счёт " + v.account},
			"; ",
		)
		found := map[vo.PIIType]bool{}
		for _, span := range detector.Detect(text) {
			found[span.Type] = true
		}
		assert.True(t, found[vo.PIISNILS] && found[vo.PIIINN] && found[vo.PIICard] && found[vo.PIIAccount],
			"case %d: checksums must pass gate 1, got %v", n, found)
	}
}
