package examples

import (
	"context"
	"errors"
	"fmt"
	ok "gook"
	"strings"
)

// RunBasicExamples demonstrates basic validation features
func RunBasicExamples() {
	ctx := context.Background()

	fmt.Println("=== Unified Rule Tree Validation Framework Examples ===")

	// Example 1: Basic String Validation
	fmt.Println("1. Basic String Validation")
	fmt.Println("-------------------------")

	stringRule := ok.All(
		ok.Test("not-empty", func(ctx context.Context, s string) error {
			if s == "" {
				return fmt.Errorf("string is empty")
			}
			return nil
		}),
		ok.Test("length", func(ctx context.Context, s string) error {
			if len(s) < 5 || len(s) > 100 {
				return fmt.Errorf("string length must be between 5 and 100")
			}
			return nil
		}),
		ok.Test("contains-at", func(ctx context.Context, s string) error {
			if !strings.Contains(s, "@") {
				return fmt.Errorf("string must contain @")
			}
			return nil
		}),
	)

	fmt.Println("Testing valid email: test@example.com")
	result, valid := stringRule.Validate(ctx, "test@example.com")
	fmt.Printf("Result: %v\n", valid)
	if !valid {
		fmt.Println(result.Format())
	}

	fmt.Println("\nTesting invalid email: ''")
	result, valid = stringRule.Validate(ctx, "")
	fmt.Printf("Result: %v\n", valid)
	if !valid {
		fmt.Println(result.Format())
	}

	// Example 2: Short-Circuit Behavior
	fmt.Println("\n\n2. Short-Circuit Behavior")
	fmt.Println("-------------------------")

	failRule1 := ok.Test("fail1", func(ctx context.Context, n int) error {
		return errors.New("first failure")
	})
	failRule2 := ok.Test("fail2", func(ctx context.Context, n int) error {
		return errors.New("second failure")
	})
	passRule := ok.Test("pass", func(ctx context.Context, n int) error {
		return nil
	})

	allRule := ok.All(failRule1, failRule2, passRule)
	fmt.Println("Testing All rule (should stop at first failure):")
	result, valid = allRule.Validate(ctx, 42)
	fmt.Printf("Result: %v\n", valid)
	fmt.Println(result.Format())

	anyRule := ok.Any(failRule1, passRule, failRule2)
	fmt.Println("\nTesting Any rule (should stop at first success):")
	result, valid = anyRule.Validate(ctx, 42)
	fmt.Printf("Result: %v\n", valid)
	fmt.Println(result.Format())

	// Example 3: Numeric Validation
	fmt.Println("\n\n3. Numeric Validation")
	fmt.Println("---------------------")

	intRule := ok.Test("range", func(ctx context.Context, n int) error {
		if n < 10 || n > 100 {
			return fmt.Errorf("value must be between 10 and 100")
		}
		return nil
	})
	fmt.Println("Testing numeric range (10-100) with value 50:")
	result, valid = intRule.Validate(ctx, 50)
	fmt.Printf("Result: %v\n", valid)
	if !valid {
		fmt.Println(result.Format())
	}

	fmt.Println("\nTesting numeric range (10-100) with value 5:")
	result, valid = intRule.Validate(ctx, 5)
	fmt.Printf("Result: %v\n", valid)
	if !valid {
		fmt.Println(result.Format())
	}

	fmt.Println("\n=== Examples Complete ===")
}

func Wip(val any) {
	ctx:= context.Background()
	rule := ok.NewRule("email",
		ok.NotNil("not-nil"),
		ok.As(ok.AssertBytes, ok.All(
			ok.BytesMax(256),
			ok.BytesMin(3),
			ok.BytesEncoding(ok.EncodingUTF8),
		)),
		ok.As(ok.AssertString, ok.All(
			ok.StringLength(3, 254),
			ok.StringContains("@"),
			ok.Any(
				ok.StringEndsWith("@jst.dev"),
				ok.StringEndsWith("@example.com"),
			),
			ok.Not(ok.Any(
				ok.StringIs("monkey@jst.dev"),
				ok.StringIs("banana@example.com"),
			)),
		)),
	)

	res, valid := rule.Validate(ctx, val)
	fmt.Printf("valid: %t\n", valid)
	fmt.Println(res.Format())
}
