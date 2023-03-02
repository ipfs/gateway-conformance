package check

import (
	"fmt"
	"regexp"
	"strings"
)

// Base Structure
// ==============

type CheckOutput struct {
	Success bool
	Reason  string
}

type Check[T any] interface {
	Check(T) CheckOutput
}

type CheckWithHint[T any] struct {
	Check Check[T]
	Hint  string
}

func WithHint[T any](hint string, check Check[T]) CheckWithHint[T] {
	return CheckWithHint[T]{
		Hint:  hint,
		Check: check,
	}
}

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

type CheckIsEqual struct {
	Value string
}

func IsEqual(value string, rest ...any) Check[string] {
	return &CheckIsEqual{
		Value: fmt.Sprintf(value, rest...),
	}
}

func (c *CheckIsEqual) Check(v string) CheckOutput {
	if v == c.Value {
		return CheckOutput{
			Success: true,
		}
	}

	return CheckOutput{
		Success: false,
		Reason:  fmt.Sprintf("expected '%s', got '%s'", c.Value, v),
	}
}

var _ Check[string] = &CheckIsEqual{}

func IsEqualWithHint(hint string, value string, rest ...any) CheckWithHint[string] {
	return WithHint(hint, IsEqual(value, rest...))
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

func (c *CheckFunc[T]) Check(v T) CheckOutput {
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
