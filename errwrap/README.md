# errwrap

```go
package main

import (
	"braces.dev/errtrace"
	"context"
	"fmt"
	"github.com/go-modulus/modulus/errlog"
	"github.com/go-modulus/modulus/errwrap"
)

func a() error { return errtrace.Errorf("failed to do a") }

func b() error { return errtrace.Errorf("failed to do b") }

func Pay(_ context.Context, paymentID string) error {
	eb := errwrap.With(
		errlog.With("paymentID", paymentID),
	)
	
	err := a()
	if err != nil {
		return errtrace.Wrap(eb(err))
    }
	return errtrace.Wrap(eb(b()))
}

func main() {
	err := Pay(context.Background(), "aboba")
	fmt.Println(errlog.Meta(err))
}

// Output:
// map[paymentId:aboba]
```