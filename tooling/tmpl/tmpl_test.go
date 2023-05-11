package tmpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicTemplating(t *testing.T) {
	x, err := fmtSafe(
		"{{first}} is {{second}} templated.",
		"this",
		"basic",
	)

	assert.Nil(t, err)
	assert.Equal(t, "this is basic templated.", x)

	x, err = fmtSafe(
		"This is a regular string.",
	)

	assert.Nil(t, err)
	assert.Equal(t, "This is a regular string.", x)

	x, err = fmtSafe(
		"{{first}} is {{second}} templated.",
		"this",
	)

	assert.NotNil(t, err)
	assert.Equal(t, "", x)

	x, err = fmtSafe(
		"{{first}} is {{second}} templated.",
		"this",
		"basic",
		"too many",
	)

	assert.NotNil(t, err)
	assert.Equal(t, "", x)
}

func TestTemplatedWrapper(t *testing.T) {
	x := Fmt(
		"{{first}} is {{second}} templated.",
		"this",
		"basic",
	)

	assert.Equal(t, "this is basic templated.", x)

	assert.Panics(t, func() {
		Fmt(
			"{{first}} is {{second}} templated.",
			"this",
		)
	})

	assert.Panics(t, func() {
		Fmt(
			"{{first}} is {{second}} templated.",
			"this",
			"basic",
			"additional",
		)
	})
}

func TestTemplatingWithReuseArguments(t *testing.T) {
	assert.Equal(t,
		"foo/foo/bar",
		Fmt(
			"{{first}}/{{first}}/{{another}}",
			"foo",
			"bar",
		),
	)

	assert.Equal(t,
		"foo/bar/foo/bar/foo/foo",
		Fmt(
			"{{first}}/{{another}}/{{first}}/{{another}}/{{first}}/{{first}}",
			"foo",
			"bar",
		),
	)

	assert.Equal(t,
		"http://Qm.ipfs.example.com/ipfs/Qm",
		Fmt(
			"{{scheme}}://{{cid}}.ipfs.{{host}}/ipfs/{{cid}}",
			"http", "Qm", "example.com",
		),
	)
}

func TestTemplatingWithEmptyNamesFails(t *testing.T) {
	v, err := fmtSafe(
		"{{first}}/{{}}/{{another}}",
		"foo",
		"bar",
		"baz",
	)

	assert.Error(t, err)
	assert.Equal(t, "", v)

	assert.Panics(t, func() {
		Fmt(
			"{{first}}/{{}}/{{another}}",
			"foo",
			"bar",
			"baz",
		)
	})
}

func TestTemplatingWithEscaping(t *testing.T) {
	assert.Equal(t,
		"{}/{{}}/{{{}}}",
		Fmt(
			"{}/{{{}}}/{{{{}}}}",
		),
	)

	assert.Equal(t,
		"{name}/{{name}}/{{{name}}}",
		Fmt(
			"{name}/{{{name}}}/{{{{name}}}}",
		),
	)

	assert.Equal(t,
		"{name}/foo/{{name}}/{{{name}}}",
		Fmt(
			"{name}/{{name}}/{{{name}}}/{{{{name}}}}",
			"foo",
		),
	)

	assert.Equal(t,
		"foo/{first}/{{}}/{{another}}/bar/{{{escaped}}}/{{first}}/baz",
		Fmt(
			"{{first}}/{first}/{{{}}}/{{{another}}}/{{}}/{{{{escaped}}}}/{{{first}}}/{{two}}",
			"foo",
			"bar",
			"baz",
		),
	)

	assert.Equal(t,
		"{{foo",
		Fmt(
			"{{{{name}}",
			"foo",
		),
	)

	assert.Equal(t,
		"foo}}",
		Fmt(
			"{{name}}}}",
			"foo",
		),
	)

	assert.Equal(t,
		"{{foo",
		Fmt(
			"{{{{}}",
			"foo",
		),
	)

	assert.Equal(t,
		"{name}}/{{foo/{barname}}}/{{{name}}}",
		Fmt(
			"{name}}/{{{{name}}/{{{}}name}}}/{{{{name}}}}",
			"foo",
			"bar",
		),
	)

}
