package go_ok

import (
	"context"
	"fmt"
)

// ThenRule represents a type-narrowing pipeline from T to U
type ThenRule[T, U any] struct {
	First     *Rule[T]
	Transform func(T) (U, error)
	Second    *Rule[U]
}

// Then creates a type-narrowing pipeline
func Then[T, U any](rule *Rule[T], transform func(T) (U, error), next *Rule[U]) *ThenRule[T, U] {
	return &ThenRule[T, U]{
		First:     rule,
		Transform: transform,
		Second:    next,
	}
}

// Validate evaluates the Then pipeline with full trace
func (tr *ThenRule[T, U]) Validate(ctx context.Context, value T) (*Result, bool) {
	result := tr.validateRecursive(ctx, value)
	return result, result.OK()
}

func (tr *ThenRule[T, U]) validateRecursive(ctx context.Context, value T) *Result {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return &Result{
			Status:  StatusFail,
			Label:   "then",
			Kind:    KindTest, // ThenRule doesn't have a RuleKind, use Test as placeholder
			Message: "context cancelled",
		}
	default:
	}

	// First validate with Rule[T]
	firstResult := tr.First.validateRecursive(ctx, value)
	if firstResult.Status == StatusFail {
		return &Result{
			Status:   StatusFail,
			Label:    "then",
			Kind:     KindTest,
			Message:  "first rule failed",
			Children: []*Result{firstResult},
		}
	}

	// Transform T -> U
	transformed, err := tr.Transform(value)
	if err != nil {
		return &Result{
			Status:   StatusFail,
			Label:    "then",
			Kind:     KindTest,
			Message:  fmt.Sprintf("transform failed: %v", err),
			Children: []*Result{firstResult},
		}
	}

	// Validate with Rule[U]
	secondResult := tr.Second.validateRecursive(ctx, transformed)

	// Combine results
	var status ResultStatus
	var message string
	if secondResult.Status == StatusPass {
		status = StatusPass
	} else if secondResult.Status == StatusFail {
		status = StatusFail
		message = "second rule failed"
	} else {
		status = StatusSkip
	}

	return &Result{
		Status:   status,
		Label:    "then",
		Kind:     KindTest,
		Message:  message,
		Children: []*Result{firstResult, secondResult},
	}
}

// ChainThen allows chaining multiple ThenRule operations
func ChainThen[T, U, V any](tr *ThenRule[T, U], transform func(U) (V, error), next *Rule[V]) *ThenRule[T, V] {
	// Create a new ThenRule that composes the transforms
	composedTransform := func(t T) (V, error) {
		u, err := tr.Transform(t)
		if err != nil {
			var zero V
			return zero, err
		}
		return transform(u)
	}

	return &ThenRule[T, V]{
		First:     tr.First,
		Transform: composedTransform,
		Second:    next,
	}
}

// String returns a human-readable representation of the ThenRule
func (tr *ThenRule[T, U]) String() string {
	return fmt.Sprintf("ThenRule[%T -> %T](%s -> %s)", 
		*new(T), *new(U), tr.First.Label, tr.Second.Label)
}
