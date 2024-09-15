# errlog

```go
package main

import (
	"braces.dev/errtrace"
	"context"
	"fmt"
	"github.com/go-modulus/modulus/errlog"
	"github.com/go-modulus/modulus/errwrap"
)

func Pay(_ context.Context, paymentID string) error {
	return errwrap.Wrap(
		errtrace.Errorf("failed to process payment"),
		errlog.With("paymentId", paymentID),
	)
}

func main() {
	err := Pay(context.Background(), "123")
	fmt.Println(errlog.Meta(err))
}

// Output:
// map[paymentId:123]
```