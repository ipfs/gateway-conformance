package tmpl

import (
	"fmt"
	"regexp"
	"strings"
)

/**
 * templatedSafe is a function that takes a format string and a variadic
 * number of arguments and returns a string with the arguments interpolated
 * into the format string.
 *
 * T("{{first}}/{{second}}/{{third}}", "foo", "bar", "baz")
 * => "foo/bar/baz"
 *
 * T("{{first}}/{{first}}/{{second}}", "foo", "bar")
 * => "foo/foo/bar"
 *
 * T("{{}}/{{}}/{{}}", "foo", "bar", "baz")
 * => "foo/bar/baz"
 *
 * T("{{first}}/{{}}/{{first}}/{{}}", "foo", "bar", "baz")
 * => "foo/bar/foo/baz"
 *
 * The format string is a Go template string, so arguments are interpolated
 * into the format string using the {{name}} syntax.
 *
 * The variables will be replaced in the order they are provided, so the
 * first argument will be interpolated into the first {{}} in the format
 * string, the second argument will be interpolated into the second {{}}.
 */
func templatedSafe(format string, args ...interface{}) (string, error) {
	re := regexp.MustCompile(`{{(\s*\w+)?\s*}}`)
	varNames := re.FindAllString(format, -1)

	data := make(map[string]interface{})
	anonymousArgs := make([]interface{}, 0)

	for _, varName := range varNames {
		name := strings.Trim(varName, "{} ")

		if len(args) == 0 {
			return "", fmt.Errorf("not enough arguments for format string: %s", format)
		}

		// you may reuse the same variable name multiple time, we use the first value.
		// {{cid}}/something/something/{{cid}}/{{suffix}}
		if _, ok := data[name]; ok {
			continue
		}

		// If the variable name is empty, we have an anonymous argument.
		if name == "" {
			anonymousArgs = append(anonymousArgs, args[0])
		} else {
			data[name] = args[0]
		}

		args = args[1:]
	}

	if len(args) > 0 {
		return "", fmt.Errorf("too many arguments for format string: %s (left: %v)", format, args)
	}

	// Apply templating
	result := re.ReplaceAllStringFunc(format, func(match string) string {
		name := strings.Trim(match, "{} ")

		// If the variable name is empty, we have an anonymous argument.
		if name == "" {
			value := anonymousArgs[0]
			anonymousArgs = anonymousArgs[1:]
			return fmt.Sprintf("%v", value)
		}

		if value, ok := data[name]; ok {
			return fmt.Sprintf("%v", value)
		}

		return match
	})

	return result, nil
}

func Templated(format string, args ...interface{}) string {
	x, err := templatedSafe(format, args...)
	if err != nil {
		panic(err)
	}

	return x
}
