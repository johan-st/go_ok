package go_ok

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestThen(t *testing.T) {
	ctx := context.Background()

	// Test successful pipeline: string -> int
	firstRule := Test("not-empty", func(ctx context.Context, s string) error {
		if s == "" {
			return errors.New("empty string")
		}
		return nil
	})

	transform := func(s string) (int, error) {
		if s == "invalid" {
			return 0, errors.New("invalid string")
		}
		var n int
		_, err := fmt.Sscanf(s, "%d", &n)
		if err != nil {
			return 0, errors.New("not a number")
		}
		return n, nil
	}

	secondRule := NumericRange(10, 100)

	pipeline := Then(firstRule, transform, secondRule)

	// Test successful validation
	result, ok := pipeline.Validate(ctx, "50")
	if !ok {
		t.Errorf("Expected validation to pass, got: %s", result.Format())
	}
	if result.Status != StatusPass {
		t.Error("Expected StatusPass")
	}
	if len(result.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(result.Children))
	}

	// Test failure at first rule
	result, ok = pipeline.Validate(ctx, "")
	if ok {
		t.Error("Expected validation to fail for empty string")
	}
	if result.Status != StatusFail {
		t.Error("Expected StatusFail")
	}
	if result.Message != "first rule failed" {
		t.Errorf("Expected 'first rule failed', got '%s'", result.Message)
	}
	if len(result.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(result.Children))
	}

	// Test failure at transform
	result, ok = pipeline.Validate(ctx, "invalid")
	if ok {
		t.Error("Expected validation to fail for invalid transform")
	}
	if result.Status != StatusFail {
		t.Error("Expected StatusFail")
	}
	if !strings.Contains(result.Message, "transform failed") {
		t.Errorf("Expected transform failure message, got '%s'", result.Message)
	}

	// Test failure at second rule (value too low)
	result, ok = pipeline.Validate(ctx, "5")
	if ok {
		t.Error("Expected validation to fail for value out of range")
	}
	if result.Status != StatusFail {
		t.Error("Expected StatusFail")
	}
	if result.Message != "second rule failed" {
		t.Errorf("Expected 'second rule failed', got '%s'", result.Message)
	}
	if len(result.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(result.Children))
	}

	// Test failure at second rule (value too high)
	result, ok = pipeline.Validate(ctx, "200")
	if ok {
		t.Error("Expected validation to fail for value out of range")
	}
	if result.Status != StatusFail {
		t.Error("Expected StatusFail")
	}
}

func TestThenContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	firstRule := Test("test", func(ctx context.Context, s string) error {
		return nil
	})
	transform := func(s string) (int, error) {
		return 50, nil
	}
	secondRule := Test("test2", func(ctx context.Context, n int) error {
		return nil
	})

	pipeline := Then(firstRule, transform, secondRule)
	result, ok := pipeline.Validate(ctx, "test")
	if ok {
		t.Error("Expected validation to fail due to context cancellation")
	}
	if result.Status != StatusFail {
		t.Error("Expected StatusFail")
	}
	if result.Message != "context cancelled" {
		t.Errorf("Expected 'context cancelled', got '%s'", result.Message)
	}
}

func TestThenWithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	firstRule := Test("slow", func(ctx context.Context, s string) error {
		time.Sleep(20 * time.Millisecond)
		return nil
	})
	transform := func(s string) (int, error) {
		return 50, nil
	}
	secondRule := Test("test", func(ctx context.Context, n int) error {
		return nil
	})

	pipeline := Then(firstRule, transform, secondRule)
	result, ok := pipeline.Validate(ctx, "test")
	if ok {
		t.Error("Expected validation to fail due to timeout")
	}
	if result.Status != StatusFail {
		t.Error("Expected StatusFail")
	}
}

func TestChainThen(t *testing.T) {
	ctx := context.Background()

	// First pipeline: string -> int
	firstRule := Test("not-empty", func(ctx context.Context, s string) error {
		if s == "" {
			return errors.New("empty")
		}
		return nil
	})
	transform1 := func(s string) (int, error) {
		var n int
		_, err := fmt.Sscanf(s, "%d", &n)
		return n, err
	}
	secondRule := NumericRange(10, 100)
	pipeline1 := Then(firstRule, transform1, secondRule)

	// Chain: int -> string
	transform2 := func(n int) (string, error) {
		return fmt.Sprintf("%d", n), nil
	}
	thirdRule := Test("length", func(ctx context.Context, s string) error {
		if len(s) < 2 {
			return errors.New("too short")
		}
		return nil
	})

	pipeline2 := ChainThen(pipeline1, transform2, thirdRule)

	// Test successful chain
	result, ok := pipeline2.Validate(ctx, "50")
	if !ok {
		t.Errorf("Expected validation to pass, got: %s", result.Format())
	}
	if result.Status != StatusPass {
		t.Error("Expected StatusPass")
	}

	// Test failure in first pipeline
	_, ok = pipeline2.Validate(ctx, "5")
	if ok {
		t.Error("Expected validation to fail for value out of range")
	}

	// Test failure in second pipeline
	_, ok = pipeline2.Validate(ctx, "1")
	if ok {
		t.Error("Expected validation to fail for single digit")
	}
}

func TestThenString(t *testing.T) {
	firstRule := Test("first", func(ctx context.Context, s string) error {
		return nil
	})
	transform := func(s string) (int, error) {
		return 0, nil
	}
	secondRule := Test("second", func(ctx context.Context, n int) error {
		return nil
	})

	pipeline := Then(firstRule, transform, secondRule)
	str := pipeline.String()
	if !strings.Contains(str, "ThenRule") {
		t.Errorf("Expected String to contain 'ThenRule', got '%s'", str)
	}
	if !strings.Contains(str, "first") {
		t.Errorf("Expected String to contain first rule label, got '%s'", str)
	}
	if !strings.Contains(str, "second") {
		t.Errorf("Expected String to contain second rule label, got '%s'", str)
	}
}

func TestThenWithComplexRules(t *testing.T) {
	ctx := context.Background()

	// Test with All rule as first
	firstRule := All(
		Test("not-empty", func(ctx context.Context, s string) error {
			if s == "" {
				return errors.New("empty")
			}
			return nil
		}),
		Test("has-at", func(ctx context.Context, s string) error {
			if !strings.Contains(s, "@") {
				return errors.New("missing @")
			}
			return nil
		}),
	)

	transform := func(s string) (int, error) {
		// Extract number after @
		parts := strings.Split(s, "@")
		if len(parts) != 2 {
			return 0, errors.New("invalid format")
		}
		var n int
		_, err := fmt.Sscanf(parts[1], "%d", &n)
		return n, err
	}

	secondRule := All(
		NumericRange(10, 100),
		Test("not-13", func(ctx context.Context, n int) error {
			if n == 13 {
				return errors.New("unlucky number")
			}
			return nil
		}),
	)

	pipeline := Then(firstRule, transform, secondRule)

	// Test success
	result, ok := pipeline.Validate(ctx, "user@50")
	if !ok {
		t.Errorf("Expected validation to pass, got: %s", result.Format())
	}

	// Test failure at first rule (missing @)
	_, ok = pipeline.Validate(ctx, "user50")
	if ok {
		t.Error("Expected validation to fail for missing @")
	}

	// Test failure at transform
	_, ok = pipeline.Validate(ctx, "user@invalid")
	if ok {
		t.Error("Expected validation to fail for invalid transform")
	}

	// Test failure at second rule (value 13)
	_, ok = pipeline.Validate(ctx, "user@13")
	if ok {
		t.Error("Expected validation to fail for unlucky number")
	}
}

