package examples

import (
	"context"
	"errors"
	"fmt"
	ok "go_ok"
)

func Numeric(testString any) {
	ctx := context.Background()
	pipeline := ok.Then(
		ok.All(ok.Test("not-nil", func(ctx context.Context, v any) error {
			if v == nil { return errors.New("required") }
			return nil
		})),
		func(v any) (int, error) {
			s, ok := v.(string)
			if !ok { return 0, errors.New("must be string") }
			var n int
			_, err := fmt.Sscanf(s, "%d", &n)
			if err != nil { return 0, errors.New("must be numeric") }
			return n, nil
		},
		ok.All(
			ok.NumericRange(10, 100),
			ok.Not(ok.NumericRange(13,13)),
		))
	result, ok := pipeline.Validate(ctx, testString)
	fmt.Printf("valid: %v, result: %+v\n", ok, result)
	fmt.Println(result.Format())
}