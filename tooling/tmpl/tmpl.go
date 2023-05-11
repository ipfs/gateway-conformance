package tmpl

import (
	"fmt"
	"regexp"
	"strings"
)

func countBraces(s string) (int, int) {
	countLeft := 0
	countRight := 0

	// Count '{' from the left
	for i := 0; i < len(s); i++ {
		if s[i] == '{' {
			countLeft++
		} else {
			break
		}
	}

	// Count '}' from the right
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '}' {
			countRight++
		} else {
			break
		}
	}

	return countLeft, countRight
}

/**
 * fmtSafe is a function that takes a format string and a variadic
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
func fmtSafe(format string, args ...interface{}) (string, error) {
	re := regexp.MustCompile(`({){2,}(\w+)?(}){2,}`)
	varNames := re.FindAllString(format, -1)

	data := make(map[string]interface{})
	anonymousArgs := make([]interface{}, 0)

	for _, varName := range varNames {
		left, right := countBraces(varName)

		min := left
		if right < left {
			min = right
		}

		if min < 2 {
			// should never happen
			return "", fmt.Errorf("invalid format string: %s", format)
		} else if min >= 3 {
			// {{{var}}} or {{{{var}}}} - we don't template this
			continue
		}
		// else, min == 2 => we do the replacement
		// Note: Even when we're too greedy and matched {{something}}}} or {{{{{something}},
		// we consume all the braces here: we are looking for the template name.
		// We add the additional braces later.
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
		left, right := countBraces(match)

		min := left
		if right < left {
			min = right
		}

		if min < 2 {
			// should never happen
			panic(fmt.Errorf("invalid format string: %s", format))
		} else if min >= 3 {
			// {{{var}}} or {{{{var}}}} - this is an escaped value.
			// we remove one brace from each side.
			return match[1 : len(match)-1]
		}

		// else, min == 2 => we do the replacement
		name := strings.Trim(match, "{} ")

		r := match

		// If the variable name is empty, we have an anonymous argument.
		if name == "" {
			value := anonymousArgs[0]
			anonymousArgs = anonymousArgs[1:]
			r = fmt.Sprintf("%v", value)
		}

		if value, ok := data[name]; ok {
			r = fmt.Sprintf("%v", value)
		}

		// add the missing braces if we've been too greedy and matched {{var}}}}}}
		if left > 2 {
			r = strings.Repeat("{", left-2) + r
		}
		if right > 2 {
			r = r + strings.Repeat("}", right-2)
		}

		return r
	})

	return result, nil
}

func Fmt(format string, args ...interface{}) string {
	x, err := fmtSafe(format, args...)
	if err != nil {
		panic(err)
	}

	return x
}
