package examples

import (
	"context"
	"fmt"
	"strings"

	ok "github.com/johan-st/gook"
)

func Email(testString string) {
	ctx := context.Background()
	emailRule := ok.All(
		ok.Test("not-empty", func(ctx context.Context, s string) error {
			if s == "" {
				return fmt.Errorf("string is empty")
			}
			return nil
		}),
		ok.Test("length", func(ctx context.Context, s string) error {
			if len(s) < 3 || len(s) > 254 {
				return fmt.Errorf("string length must be between 3 and 254")
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
	res, valid := emailRule.Validate(ctx, testString)
	fmt.Printf("valid: %v\n", valid)
	fmt.Println(res.Format())
}