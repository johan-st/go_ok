package examples

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	ok "github.com/johan-st/gook"
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
	ctx := context.Background()
	rule := ok.NewRule("email",
		ok.NotNil("not-nil"),
		ok.Test("bytes-validation", func(ctx context.Context, v any) error {
			var bytes []byte
			switch v := v.(type) {
			case []byte:
				bytes = v
			case string:
				bytes = []byte(v)
			default:
				return fmt.Errorf("value is not []byte or string")
			}
			
			if len(bytes) > 256 {
				return fmt.Errorf("bytes too long (max: 256, got: %d)", len(bytes))
			}
			if len(bytes) < 3 {
				return fmt.Errorf("bytes too short (min: 3, got: %d)", len(bytes))
			}
			if !utf8.Valid(bytes) {
				return fmt.Errorf("bytes are not valid UTF-8")
			}
			return nil
		}),
		ok.Test("string-validation", func(ctx context.Context, v any) error {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("value is not a string")
			}
			
			if len(s) < 3 || len(s) > 254 {
				return fmt.Errorf("string length must be between 3 and 254")
			}
			if !strings.Contains(s, "@") {
				return fmt.Errorf("string does not contain @")
			}
			if !strings.HasSuffix(s, "@jst.dev") && !strings.HasSuffix(s, "@example.com") {
				return fmt.Errorf("string does not end with @jst.dev or @example.com")
			}
			if s == "monkey@jst.dev" || s == "banana@example.com" {
				return fmt.Errorf("string is not allowed")
			}
			return nil
		}),
	)

	res, valid := rule.Validate(ctx, val)
	fmt.Printf("valid: %t\n", valid)
	fmt.Println(res.Format())
}
