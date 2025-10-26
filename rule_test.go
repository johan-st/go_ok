package go_ok

import (
	"context"
	"errors"
	"testing"
)

func TestBasicValidation(t *testing.T) {
	ctx := context.Background()

	// Test basic string validation
	stringRule := All(
		NotEmpty("email"),
		StringLength(5, 100),
		Test("email-format", func(ctx context.Context, s string) error {
			if s == "" {
				return errors.New("empty")
			}
			return nil
		}),
	)

	result, ok := stringRule.Validate(ctx, "test@example.com")
	if !ok {
		t.Errorf("Expected validation to pass, got: %s", result.Format())
	}

	result, ok = stringRule.Validate(ctx, "")
	if ok {
		t.Error("Expected validation to fail for empty string")
	}
}

func TestShortCircuit(t *testing.T) {
	ctx := context.Background()

	failRule1 := Test("fail1", func(ctx context.Context, n int) error {
		return errors.New("first failure")
	})
	failRule2 := Test("fail2", func(ctx context.Context, n int) error {
		return errors.New("second failure")
	})
	passRule := Test("pass", func(ctx context.Context, n int) error {
		return nil
	})

	allRule := All(failRule1, failRule2, passRule)
	result, ok := allRule.Validate(ctx, 42)
	if ok {
		t.Error("Expected All rule to fail")
	}

	// Check that only first rule failed, others are skipped
	if len(result.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(result.Children))
	}
	if result.Children[0].Status != StatusFail {
		t.Error("Expected first child to fail")
	}
	if result.Children[1].Status != StatusSkip {
		t.Error("Expected second child to be skipped")
	}
	if result.Children[2].Status != StatusSkip {
		t.Error("Expected third child to be skipped")
	}
}
