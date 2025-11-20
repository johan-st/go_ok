package gook

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// RuleKind represents the type of rule node
type RuleKind int

const (
	KindTest RuleKind = iota
	KindAll
	KindAny
	KindNot
)

// String returns a human-readable representation of the rule kind
func (k RuleKind) String() string {
	switch k {
	case KindTest:
		return "test"
	case KindAll:
		return "all"
	case KindAny:
		return "any"
	case KindNot:
		return "not"
	default:
		return "unknown"
	}
}

// Rule represents a validation rule for type T
// Rules are immutable and thread-safe
type Rule[T any] struct {
	Label    string
	Kind     RuleKind
	TestFn   func(context.Context, T) error // returns error for message
	Children []*Rule[T]                     // only same-typed children
}

// Test creates a leaf test rule
func Test[T any](label string, fn func(context.Context, T) error) *Rule[T] {
	return &Rule[T]{
		Label:  label,
		Kind:   KindTest,
		TestFn: fn,
	}
}

// All creates an AND combinator that stops at first failure
func All[T any](rules ...*Rule[T]) *Rule[T] {
	return &Rule[T]{
		Label:    "all",
		Kind:     KindAll,
		Children: rules,
	}
}

// Any creates an OR combinator that stops at first success
func Any[T any](rules ...*Rule[T]) *Rule[T] {
	return &Rule[T]{
		Label:    "any",
		Kind:     KindAny,
		Children: rules,
	}
}

// Not creates a negation rule
func Not[T any](rule *Rule[T]) *Rule[T] {
	return &Rule[T]{
		Label:    "not",
		Kind:     KindNot,
		Children: []*Rule[T]{rule},
	}
}

// NewRule creates a labeled rule that combines multiple rules with All combinator
func NewRule(label string, rules ...*Rule[any]) *Rule[any] {
	return &Rule[any]{
		Label:    label,
		Kind:     KindAll,
		Children: rules,
	}
}

// As creates a type-narrowing/transformation rule from any to T
// If the transform fails, the As rule fails
func As[T any](transformFn func(any) (T, error), rule *Rule[T]) *Rule[any] {
	return Test("as", func(ctx context.Context, val any) error {
		// Transform the value
		transformed, err := transformFn(val)
		if err != nil {
			return fmt.Errorf("transform failed: %v", err)
		}
		
		// Validate with the rule
		result, valid := rule.Validate(ctx, transformed)
		if !valid {
			return fmt.Errorf("validation failed: %s", result.Message)
		}
		
		return nil
	})
}


// Validate evaluates the rule against the given value with full trace
func (r *Rule[T]) Validate(ctx context.Context, value T) (*Result, bool) {
	result := r.validateRecursive(ctx, value)
	return result, result.OK()
}

func (r *Rule[T]) validateRecursive(ctx context.Context, value T) *Result {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return &Result{
			Status:  StatusFail,
			Label:   r.Label,
			Kind:    r.Kind,
			Message: "context cancelled",
		}
	default:
	}

	switch r.Kind {
	case KindTest:
		return r.validateTest(ctx, value)
	case KindAll:
		return r.validateAll(ctx, value)
	case KindAny:
		return r.validateAny(ctx, value)
	case KindNot:
		return r.validateNot(ctx, value)
	default:
		return &Result{
			Status:  StatusFail,
			Label:   r.Label,
			Kind:    r.Kind,
			Message: fmt.Sprintf("unknown rule kind: %v", r.Kind),
		}
	}
}

func (r *Rule[T]) validateTest(ctx context.Context, value T) *Result {
	if err := r.TestFn(ctx, value); err != nil {
		return &Result{
			Status:  StatusFail,
			Label:   r.Label,
			Kind:    KindTest,
			Message: err.Error(),
		}
	}
	return &Result{
		Status: StatusPass,
		Label:  r.Label,
		Kind:   KindTest,
	}
}

func (r *Rule[T]) validateAll(ctx context.Context, value T) *Result {
	children := make([]*Result, len(r.Children))

	for i, child := range r.Children {
		childResult := child.validateRecursive(ctx, value)
		children[i] = childResult

		// Short-circuit on first failure
		if childResult.Status == StatusFail {
			// Mark remaining children as skipped
			for j := i + 1; j < len(r.Children); j++ {
				children[j] = &Result{
					Status: StatusSkip,
					Label:  r.Children[j].Label,
					Kind:   r.Children[j].Kind,
				}
			}
			return &Result{
				Status:   StatusFail,
				Label:    r.Label,
				Kind:     KindAll,
				Children: children,
			}
		}
	}

	return &Result{
		Status:   StatusPass,
		Label:    r.Label,
		Kind:     KindAll,
		Children: children,
	}
}

func (r *Rule[T]) validateAny(ctx context.Context, value T) *Result {
	children := make([]*Result, len(r.Children))

	for i, child := range r.Children {
		childResult := child.validateRecursive(ctx, value)
		children[i] = childResult

		// Short-circuit on first success
		if childResult.Status == StatusPass {
			// Mark remaining children as skipped
			for j := i + 1; j < len(r.Children); j++ {
				children[j] = &Result{
					Status: StatusSkip,
					Label:  r.Children[j].Label,
					Kind:   r.Children[j].Kind,
				}
			}
			return &Result{
				Status:   StatusPass,
				Label:    r.Label,
				Kind:     KindAny,
				Children: children,
			}
		}
	}

	return &Result{
		Status:   StatusFail,
		Label:    r.Label,
		Kind:     KindAny,
		Children: children,
	}
}

func (r *Rule[T]) validateNot(ctx context.Context, value T) *Result {
	if len(r.Children) != 1 {
		return &Result{
			Status:  StatusFail,
			Label:   r.Label,
			Kind:    KindNot,
			Message: "not rule must have exactly one child",
		}
	}

	childResult := r.Children[0].validateRecursive(ctx, value)

	// Invert the result
	var status ResultStatus
	var message string
	if childResult.Status == StatusPass {
		status = StatusFail
		message = "not rule failed (child passed)"
	} else if childResult.Status == StatusFail {
		status = StatusPass
	} else {
		status = StatusSkip
	}

	return &Result{
		Status:   status,
		Label:    r.Label,
		Kind:     KindNot,
		Message:  message,
		Children: []*Result{childResult},
	}
}

// OneOf creates a rule that passes if exactly one of the given rules passes
func OneOf[T any](rules ...*Rule[T]) *Rule[T] {
	return Test("one-of", func(ctx context.Context, value T) error {
		passCount := 0
		var lastError error

		for _, rule := range rules {
			result, _ := rule.Validate(ctx, value)
			if result.OK() {
				passCount++
			} else {
				lastError = errors.New(result.Message)
			}
		}

		if passCount == 0 {
			return fmt.Errorf("none of the rules passed: %v", lastError)
		} else if passCount > 1 {
			return errors.New("multiple rules passed (expected exactly one)")
		}
		return nil
	})
}

// NotNil creates a rule that ensures a value is not nil
func NotNil(label string) *Rule[any] {
	return Test(label, func(ctx context.Context, value any) error {
		if value == nil {
			return errors.New("value is nil")
		}
		return nil
	})
}

// AssertBytes is a transform function that converts any to []byte
func AssertBytes(v any) ([]byte, error) {
	switch v := v.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("value is not []byte or string")
	}
}

// AssertString is a transform function that converts any to string
func AssertString(v any) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("value is not a string")
	}
	return s, nil
}

// BytesMax creates a rule for maximum byte length
func BytesMax(max int) *Rule[[]byte] {
	return Test("bytes-max", func(ctx context.Context, value []byte) error {
		if len(value) > max {
			return fmt.Errorf("bytes too long (max: %d, got: %d)", max, len(value))
		}
		return nil
	})
}

// BytesMin creates a rule for minimum byte length
func BytesMin(min int) *Rule[[]byte] {
	return Test("bytes-min", func(ctx context.Context, value []byte) error {
		if len(value) < min {
			return fmt.Errorf("bytes too short (min: %d, got: %d)", min, len(value))
		}
		return nil
	})
}

// Encoding represents text encoding types
type Encoding int

const (
	EncodingUTF8 Encoding = iota
)

// BytesEncoding creates a rule for byte encoding validation
func BytesEncoding(enc Encoding) *Rule[[]byte] {
	return Test("bytes-encoding", func(ctx context.Context, value []byte) error {
		// Basic encoding check - can be enhanced later
		switch enc {
		case EncodingUTF8:
			if !utf8.Valid(value) {
				return fmt.Errorf("bytes are not valid UTF-8")
			}
		default:
			return errors.New("unknown encoding")
		}

		return nil
	})
}

// StringLength creates a rule for string length validation
func StringLength(min, max int) *Rule[string] {
	return Test("string-length", func(ctx context.Context, value string) error {
		length := len(value)
		if length < min {
			return fmt.Errorf("string too short (min: %d, got: %d)", min, length)
		}
		if length > max {
			return fmt.Errorf("string too long (max: %d, got: %d)", max, length)
		}
		return nil
	})
}

// StringContains creates a rule that checks if a string contains a substring
func StringContains(substring string) *Rule[string] {
	return Test("string-contains", func(ctx context.Context, value string) error {
		if !strings.Contains(value, substring) {
			return fmt.Errorf("string does not contain %s", substring)
		}
		return nil
	})
}

// StringEndsWith creates a rule that checks if a string ends with a suffix
func StringEndsWith(suffix string) *Rule[string] {
	return Test("string-ends-with", func(ctx context.Context, value string) error {
		if !strings.HasSuffix(value, suffix) {
			return fmt.Errorf("string does not end with %s", suffix)
		}
		return nil
	})
}

// StringIs creates a rule that checks if a string equals a value
func StringIs(value string) *Rule[string] {
	return Test("string-is", func(ctx context.Context, s string) error {
		if s != value {
			return fmt.Errorf("string is not %s", value)
		}
		return nil
	})
}
