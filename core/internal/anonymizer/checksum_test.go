package anonymizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChecksums(t *testing.T) {
	t.Parallel()

	assert.True(t, luhnValid("4111111111111111"))
	assert.False(t, luhnValid("4111111111111112"))
	assert.False(t, luhnValid("1234"))

	assert.True(t, snilsValid("11223344595"))
	assert.False(t, snilsValid("11223344596"))
	assert.False(t, snilsValid("112233445"))

	assert.True(t, innValid("7707083893"))
	assert.False(t, innValid("7707083894"))
	assert.True(t, innValid("500100732259"))
	assert.False(t, innValid("500100732258"))
	assert.False(t, innValid("12345"))
}
