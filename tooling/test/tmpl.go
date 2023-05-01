package test

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

/**
 * templatedSafe is a function that takes a format string and a variadic
 * number of arguments and returns a string with the arguments interpolated
 * into the format string.
 *
 * The format string is a Go template string, so arguments are interpolated
 * into the format string using the {{.arg}} syntax.
 *
 * The format string may contain the special string "{{}}" which will be
 * replaced with "{{.arg0}}" for the first argument, "{{.arg1}}" for the
 * second argument, and so on.
 */
func templatedSafe(format string, args ...interface{}) (string, error) {
	if strings.Count(format, "{{}}") != len(args) {
		return "", fmt.Errorf(
			"format string contains %d '{{}}' but %d arguments were provided",
			strings.Count(format, "{{}}"),
			len(args),
		)
	}

	for i := 0; i < len(args); i++ {
		format = strings.Replace(format, "{{}}", fmt.Sprintf("{{.arg%d}}", i), 1)
	}

	tmpl, err := template.New("tmpl").Parse(format)
	if err != nil {
		return "", err
	}

	data := make(map[string]interface{})
	for i, arg := range args {
		data[fmt.Sprintf("arg%d", i)] = arg
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func Templated(format string, args ...interface{}) string {
	x, err := templatedSafe(format, args...)
	if err != nil {
		panic(err)
	}

	return x
}