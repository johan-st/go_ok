package gook

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

// ThenRuleBuilder provides fluent chaining for Rule construction with type narrowing
// Now returns *Rule[T] instead of *ThenRule[T, U]
type ThenRuleBuilder[T, U any] struct {
	first     *Rule[T]
	transform func(T) (U, error)
}

// All creates a Rule with All combinator on the second rule
func (trb *ThenRuleBuilder[T, U]) All(rules ...*Rule[U]) *Rule[T] {
	return Then(trb.first, trb.transform, All(rules...))
}

// Any creates a Rule with Any combinator on the second rule
func (trb *ThenRuleBuilder[T, U]) Any(rules ...*Rule[U]) *Rule[T] {
	return Then(trb.first, trb.transform, Any(rules...))
}

// Rule creates a Rule with a single rule
func (trb *ThenRuleBuilder[T, U]) Rule(rule *Rule[U]) *Rule[T] {
	return Then(trb.first, trb.transform, rule)
}

// Then is now defined in rule.go and returns *Rule[T] instead of *ThenRule[T, U]
// This file maintains ThenRule for backward compatibility during migration

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

// All creates a new ThenRule with All combinator on the second rule
func (tr *ThenRule[T, U]) All(rules ...*Rule[U]) *ThenRule[T, U] {
	return &ThenRule[T, U]{
		First:     tr.First,
		Transform: tr.Transform,
		Second:    All(rules...),
	}
}

// Any creates a new ThenRule with Any combinator on the second rule
func (tr *ThenRule[T, U]) Any(rules ...*Rule[U]) *ThenRule[T, U] {
	return &ThenRule[T, U]{
		First:     tr.First,
		Transform: tr.Transform,
		Second:    Any(rules...),
	}
}

// ChainThen allows chaining multiple Then operations
// Now works with Rule[T] instead of ThenRule[T, U]
func ChainThen[T, U, V any](first *Rule[T], firstTransform func(T) (U, error), secondTransform func(U) (V, error), next *Rule[V]) *Rule[T] {
	// Create a composed transform T -> V
	composedTransform := func(t T) (V, error) {
		u, err := firstTransform(t)
		if err != nil {
			var zero V
			return zero, err
		}
		return secondTransform(u)
	}

	return Then(first, composedTransform, next)
}

// ThenString is a convenience helper for the common pattern of validating any -> string
// It combines NotNil check, AsString transform, and string rules
// Now returns *Rule[any] instead of *ThenRule[any, string]
func ThenString(first *Rule[any], rules ...*Rule[string]) *Rule[any] {
	return Then(first, AsString, All(rules...))
}

// String returns a human-readable representation of the ThenRule
func (tr *ThenRule[T, U]) String() string {
	return fmt.Sprintf("ThenRule[%T -> %T](%s -> %s)", 
		*new(T), *new(U), tr.First.Label, tr.Second.Label)
}
