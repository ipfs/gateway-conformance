package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicTemplating(t *testing.T) {
	x, err := templatedSafe(
		"{{}} is {{}} templated.",
		"this",
		"basic",
	)

	assert.Nil(t, err)
	assert.Equal(t, "this is basic templated.", x)

	x, err = templatedSafe(
		"{{}} is {{}} templated.",
		"this",
	)

	assert.NotNil(t, err)
	assert.Equal(t, "", x)

	x, err = templatedSafe(
		"{{}} is {{}} templated.",
		"this",
		"basic",
		"too many",
	)

	assert.NotNil(t, err)
	assert.Equal(t, "", x)
}

func TestTemplatedWrapper(t *testing.T) {
	x := Templated(
		"{{}} is {{}} templated.",
		"this",
		"basic",
	)

	assert.Equal(t, "this is basic templated.", x)

	assert.Panics(t, func() {
		Templated(
			"{{}} is {{}} templated.",
			"this",
		)
	})

	assert.Panics(t, func() {
		Templated(
			"{{}} is {{}} templated.",
			"this",
			"basic",
			"additional",
		)
	})
}
