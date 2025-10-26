package examples

import (
	"context"
	"errors"
	"fmt"
	"go_ok"
	"strings"
)

// RunBasicExamples demonstrates basic validation features
func RunBasicExamples() {
	ctx := context.Background()

	fmt.Println("=== Unified Rule Tree Validation Framework Examples ===")

	// Example 1: Basic String Validation
	fmt.Println("1. Basic String Validation")
	fmt.Println("-------------------------")
	
	stringRule := go_ok.All(
		go_ok.NotEmpty("email"),
		go_ok.StringLength(5, 100),
		go_ok.Test("email-format", func(ctx context.Context, s string) error {
			if !strings.Contains(s, "@") {
				return errors.New("invalid email format")
			}
			return nil
		}),
	)

	fmt.Println("Testing valid email: test@example.com")
	result, ok := stringRule.Validate(ctx, "test@example.com")
	fmt.Printf("Result: %v\n", ok)
	if !ok {
		fmt.Println(result.Format())
	}

	fmt.Println("\nTesting invalid email: ''")
	result, ok = stringRule.Validate(ctx, "")
	fmt.Printf("Result: %v\n", ok)
	if !ok {
		fmt.Println(result.Format())
	}

	// Example 2: Short-Circuit Behavior
	fmt.Println("\n\n2. Short-Circuit Behavior")
	fmt.Println("-------------------------")
	
	failRule1 := go_ok.Test("fail1", func(ctx context.Context, n int) error {
		return errors.New("first failure")
	})
	failRule2 := go_ok.Test("fail2", func(ctx context.Context, n int) error {
		return errors.New("second failure")
	})
	passRule := go_ok.Test("pass", func(ctx context.Context, n int) error {
		return nil
	})

	allRule := go_ok.All(failRule1, failRule2, passRule)
	fmt.Println("Testing All rule (should stop at first failure):")
	result, ok = allRule.Validate(ctx, 42)
	fmt.Printf("Result: %v\n", ok)
	fmt.Println(result.Format())

	anyRule := go_ok.Any(failRule1, passRule, failRule2)
	fmt.Println("\nTesting Any rule (should stop at first success):")
	result, ok = anyRule.Validate(ctx, 42)
	fmt.Printf("Result: %v\n", ok)
	fmt.Println(result.Format())

	// Example 3: Numeric Validation
	fmt.Println("\n\n3. Numeric Validation")
	fmt.Println("---------------------")
	
	intRule := go_ok.NumericRange(10, 100)
	fmt.Println("Testing numeric range (10-100) with value 50:")
	result, ok = intRule.Validate(ctx, 50)
	fmt.Printf("Result: %v\n", ok)
	if !ok {
		fmt.Println(result.Format())
	}

	fmt.Println("\nTesting numeric range (10-100) with value 5:")
	result, ok = intRule.Validate(ctx, 5)
	fmt.Printf("Result: %v\n", ok)
	if !ok {
		fmt.Println(result.Format())
	}

	// Example 4: Required/Optional
	fmt.Println("\n\n4. Required/Optional")
	fmt.Println("--------------------")
	
	requiredRule := go_ok.Required[string]("name")
	fmt.Println("Testing Required with empty string:")
	result, ok = requiredRule.Validate(ctx, "")
	fmt.Printf("Result: %v\n", ok)
	if !ok {
		fmt.Println(result.Format())
	}

	fmt.Println("\nTesting Required with non-empty string:")
	result, ok = requiredRule.Validate(ctx, "John")
	fmt.Printf("Result: %v\n", ok)
	if !ok {
		fmt.Println(result.Format())
	}

	optionalRule := go_ok.Optional(go_ok.NotEmpty("email"))
	fmt.Println("\nTesting Optional with empty string:")
	result, ok = optionalRule.Validate(ctx, "")
	fmt.Printf("Result: %v\n", ok)
	if !ok {
		fmt.Println(result.Format())
	}

	fmt.Println("\nTesting Optional with valid email:")
	result, ok = optionalRule.Validate(ctx, "test@example.com")
	fmt.Printf("Result: %v\n", ok)
	if !ok {
		fmt.Println(result.Format())
	}

	fmt.Println("\n=== Examples Complete ===")
}
