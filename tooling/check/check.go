package check

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
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

func (c CheckIsEmpty) Check(v string) CheckOutput {
	if v == "" {
		return CheckOutput{
			Success: true,
		}
	}

	return CheckOutput{
		Success: false,
		Reason:  fmt.Sprintf("expected empty string, got '%s'", v),
	}
}

var _ Check[string] = CheckIsEmpty{}

func IsEmpty(hint ...string) interface{} {
	if len(hint) > 1 {
		panic("hint can only be one string")
	}
	if len(hint) == 1 {
		return WithHint[string](
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
		Value: fmt.Sprintf(value, rest...),
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

	return CheckOutput{
		Success: false,
		Reason:  fmt.Sprintf("expected '%v', got '%v'", c.Value, v),
	}
}

var _ Check[[]byte] = CheckIsEqualBytes{}

func IsEqualWithHint(hint string, value string, rest ...any) CheckWithHint[string] {
	return WithHint[string](hint, IsEqual(value, rest...))
}

type CheckContains struct {
	Value string
}

func Contains(value string, rest ...any) Check[string] {
	return &CheckContains{
		Value: fmt.Sprintf(value, rest...),
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
	str := fmt.Sprintf(value, rest...)

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

type CheckNot struct {
	check Check[string]
}

func Not(check Check[string]) Check[string] {
	return CheckNot{
		check: check,
	}
}

func (c CheckNot) Check(v string) CheckOutput {
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

var _ Check[string] = CheckNot{}

type CheckIsJSONEqual struct {
	Value interface{}
}

func IsJSONEqual(value []byte) Check[[]byte] {
	var result interface{}
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
		return CheckOutput{
			Success: false,
			Reason:  fmt.Sprintf("expected '%s' to be valid JSON, but got error: %s", string(v), err),
		}
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
