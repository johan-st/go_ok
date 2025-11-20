package examples

import (
	"context"
	"errors"
	"fmt"

	ok "github.com/johan-st/gook"
)

func Numeric(testString any) {
	ctx := context.Background()
	
	// Transform and validate as int
	rule := ok.Test("numeric-validation", func(ctx context.Context, v any) error {
		// Check not nil
		if v == nil {
			return errors.New("required")
		}
		
		// Transform to string
		s, ok := v.(string)
		if !ok {
			return errors.New("must be string")
		}
		
		// Transform to int
		var n int
		_, err := fmt.Sscanf(s, "%d", &n)
		if err != nil {
			return errors.New("must be numeric")
		}
		
		// Validate range
		if n < 10 || n > 100 {
			return fmt.Errorf("value must be between 10 and 100")
		}
		
		// Validate not 13
		if n == 13 {
			return fmt.Errorf("value is not 13")
		}
		
		return nil
	})
	
	result, valid := rule.Validate(ctx, testString)
	fmt.Printf("valid: %v\n", valid)
	fmt.Println(result.Format())
}