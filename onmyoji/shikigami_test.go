package onmyoji

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetShikigami(t *testing.T) {
	shiki, err := GetShikigami("iba")
	assert.NoError(t, err)
	assert.Equal(t, 3216, shiki.Atk)

	shiki, err = GetShikigami("Ibaraki doji")
	assert.NoError(t, err)
	assert.Equal(t, 3216, shiki.Atk)

	shiki, err = GetShikigami("ibara")
	assert.Error(t, err)
}
