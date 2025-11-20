package examples

import (
	"context"
	"fmt"
	ok "gook"
)

func Email(testString string) {
	ctx := context.Background()
	emailRule := ok.All(
		ok.Not(ok.StringEmpty()),
		ok.StringLength(3, 254),
		ok.StringContains("@"),
	)
	res, valid := emailRule.Validate(ctx, testString)
	fmt.Printf("valid: %v, result:\n%s", valid, res.Format())
	fmt.Println(res.Format())
}