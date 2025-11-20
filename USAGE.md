```go
package main

import (
	"context"
	"fmt"
	ok "gook"
)

func main() {
	ctx := context.Background()
    value := getAnyValue()

    rule := ok.NewRule("email", 
        ok.NotNil,
        ok.As(ok.AssertBytes, ok.All(
            ok.BytesMax(256),
            ok.BytesMin(3),
            ok.OneOf(
                ok.BytesEncoding(ok.EncodingUTF8),
                ok.BytesEncoding(ok.EncodingUTF16),
                ok.BytesEncoding(ok.EncodingUTF32),
            ),
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

	res, valid := rule.Validate(ctx, value)
	fmt.Printf("valid: %t\n", valid)
	fmt.Println(res.Format())
}
```