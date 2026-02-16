package check

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/ipfs/gateway-conformance/tooling/tmpl"
)

// Base Structure
// ==============

type CheckOutput struct {
	Success bool
	Reason  string
	Err     error
	Hint    string
}

type Check[T any] interface {
	Check(T) CheckOutput
}

type CheckWithHint[T any] struct {
	Check_ Check[T]
	Hint   string
}

func WithHint[T any](hint string, check Check[T]) CheckWithHint[T] {
	return CheckWithHint[T]{
		Hint:   hint,
		Check_: check,
	}
}

func (c CheckWithHint[T]) Check(v T) CheckOutput {
	output := c.Check_.Check(v)

	if output.Hint == "" {
		output.Hint = c.Hint
	} else {
		output.Hint = fmt.Sprintf("%s (%s)", c.Hint, output.Hint)
	}

	return output
}

var _ Check[string] = CheckWithHint[string]{}

// Base
// ====

type CheckIsEmpty struct {
}

func (c CheckIsEmpty) Check(v []string) CheckOutput {
	if len(v) == 0 {
		return CheckOutput{
			Success: true,
		}
	}

	return CheckOutput{
		Success: false,
		Reason:  fmt.Sprintf("expected empty array, got '%s'", v),
	}
}

var _ Check[[]string] = CheckIsEmpty{}

func IsEmpty(hint ...string) any {
	if len(hint) > 1 {
		panic("hint can only be one string")
	}
	if len(hint) == 1 {
		return WithHint[[]string](
			hint[0],
			CheckIsEmpty{},
		)
	}
	return CheckIsEmpty{}
}

type CheckAnd[T any] struct {
	Checks []Check[T]
}

func And[T any](checks ...Check[T]) Check[T] {
	return &CheckAnd[T]{
		Checks: checks,
	}
}

func (c *CheckAnd[T]) Check(v T) CheckOutput {
	for _, check := range c.Checks {
		output := check.Check(v)
		if !output.Success {
			return output
		}
	}

	return CheckOutput{
		Success: true,
	}
}

type CheckIsEqual[T comparable] struct {
	Value T
}

func IsEqual(value string, rest ...any) CheckIsEqual[string] {
	return CheckIsEqual[string]{
		Value: tmpl.Fmt(value, rest...),
	}
}

func IsEqualT[T comparable](value T) *CheckIsEqual[T] {
	return &CheckIsEqual[T]{
		Value: value,
	}
}

func (c CheckIsEqual[T]) Check(v T) CheckOutput {
	if v == c.Value {
		return CheckOutput{
			Success: true,
		}
	}

	return CheckOutput{
		Success: false,
		Reason:  fmt.Sprintf("expected '%v', got '%v'", c.Value, v),
	}
}

var _ Check[string] = CheckIsEqual[string]{}

type CheckIsEqualBytes struct {
	Value []byte
}

// golang doesn't support method overloading / generic specialization
func IsEqualBytes(value []byte) Check[[]byte] {
	return CheckIsEqualBytes{
		Value: value,
	}
}

func (c CheckIsEqualBytes) Check(v []byte) CheckOutput {
	if bytes.Equal(v, c.Value) {
		return CheckOutput{
			Success: true,
		}
	}

	var reason string
	if utf8.Valid(v) && utf8.Valid(c.Value) {
		// Print human-readable plain text, when possible
		reason = fmt.Sprintf("expected %q, got %q", c.Value, v)
	} else {
		// Print byte codes
		reason = fmt.Sprintf("expected '%v', got '%v'", c.Value, v)
	}

	return CheckOutput{
		Success: false,
		Reason:  reason,
	}
}

var _ Check[[]byte] = CheckIsEqualBytes{}

func IsEqualWithHint(hint string, value string, rest ...any) CheckWithHint[string] {
	return WithHint[string](hint, IsEqual(value, rest...))
}

type CheckUniqAnd struct {
	check Check[string]
}

var _ Check[[]string] = &CheckUniqAnd{}

func IsUniqAnd(check Check[string]) Check[[]string] {
	return &CheckUniqAnd{
		check: check,
	}
}

func (c *CheckUniqAnd) Check(v []string) CheckOutput {
	if len(v) != 1 {
		return CheckOutput{
			Success: false,
			Reason:  "expected one element",
		}
	}

	return c.check.Check(v[0])
}

type CheckHas struct {
	values []string
}

var _ Check[[]string] = &CheckHas{}

func Has(values ...string) Check[[]string] {
	return &CheckHas{
		values: values,
	}
}

func (c *CheckHas) Check(v []string) CheckOutput {
	for _, value := range c.values {
		found := slices.Contains(v, value)

		if !found {
			return CheckOutput{
				Success: false,
				Reason:  fmt.Sprintf("expected to find '%s' in '%s'", value, v),
			}
		}
	}

	return CheckOutput{
		Success: true,
	}
}

type CheckContains struct {
	Value string
}

func Contains(value string, rest ...any) Check[string] {
	return &CheckContains{
		Value: tmpl.Fmt(value, rest...),
	}
}

func (c *CheckContains) Check(v string) CheckOutput {
	if strings.Contains(v, c.Value) {
		return CheckOutput{
			Success: true,
		}
	}

	return CheckOutput{
		Success: false,
		Reason:  fmt.Sprintf("expect to find substring '%s', got '%s'", c.Value, v),
	}
}

func ContainsWithHint(hint string, value string, rest ...any) CheckWithHint[string] {
	return WithHint(hint, Contains(value, rest...))
}

var _ Check[string] = &CheckContains{}

type CheckRegexpMatch struct {
	Value *regexp.Regexp
}

func (c *CheckRegexpMatch) Check(v string) CheckOutput {
	if c.Value.MatchString(v) {
		return CheckOutput{
			Success: true,
		}
	}

	return CheckOutput{
		Success: false,
		Reason:  fmt.Sprintf("expected to match '%s', got '%s'", c.Value.String(), v),
	}
}

func Matches(value string, rest ...any) Check[string] {
	str := tmpl.Fmt(value, rest...)

	return &CheckRegexpMatch{
		Value: regexp.MustCompile(str),
	}
}

var _ Check[string] = &CheckRegexpMatch{}

type CheckFunc[T any] struct {
	Fn func(T) bool
}

func Checks[T any](hint string, f func(T) bool) CheckWithHint[T] {
	return WithHint[T](hint, &CheckFunc[T]{
		Fn: f,
	})
}

func (c CheckFunc[T]) Check(v T) CheckOutput {
	if c.Fn(v) {
		return CheckOutput{
			Success: true,
		}
	}

	return CheckOutput{
		Success: false,
		Reason:  fmt.Sprintf("expected to 'f(%v) = true'", v),
	}
}

var _ Check[string] = &CheckFunc[string]{}

type CheckNot[T any] struct {
	check Check[T]
}

func Not[T any](check Check[T]) Check[T] {
	return CheckNot[T]{
		check: check,
	}
}

func (c CheckNot[T]) Check(v T) CheckOutput {
	result := c.check.Check(v)

	if result.Success {
		return CheckOutput{
			Success: false,
			Reason:  fmt.Sprintf("expected %v to fail, but it succeeded", c.check),
		}
	}

	return CheckOutput{
		Success: true,
	}
}

var _ Check[string] = CheckNot[string]{}

type CheckIsJSONEqual struct {
	Value any
}

func IsJSONEqual(value []byte) Check[[]byte] {
	var result any
	err := json.Unmarshal(value, &result)
	if err != nil {
		panic(err) // TODO: move a t.Testing around to `t.Fatal` this case
	}

	return &CheckIsJSONEqual{
		Value: result,
	}
}

func (c *CheckIsJSONEqual) Check(v []byte) CheckOutput {
	var o map[string]any
	err := json.Unmarshal(v, &o)
	if err != nil {
		panic(err) // TODO: move a t.Testing around to call `t.Fatal` on this case
	}

	if reflect.DeepEqual(o, c.Value) {
		return CheckOutput{
			Success: true,
		}
	}

	b, err := json.Marshal(c.Value)
	if err != nil {
		panic(err) // TODO: move a t.Testing around to call `t.Fatal` on this case
	}

	return CheckOutput{
		Success: false,
		Reason:  fmt.Sprintf("expected '%s', got '%s'", string(b), string(v)),
	}
}

var _ Check[[]byte] = &CheckIsJSONEqual{}
