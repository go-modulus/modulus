package translation

import (
	"context"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/http/errhttp"
)

func LocalizeErrorHint() errhttp.ErrorProcessor {
	return func(ctx context.Context, err error) error {
		if err == nil {
			return nil
		}
		localizer, errL := GetLocalizer(ctx)
		if errL != nil {
			// If we cannot get the localizer, we just return the original error
			return err
		}
		if localizer != nil {
			hint := errors.Hint(err)
			domain := Domain(err)
			args := HintArguments(err)
			if domain != "" {
				if len(args) > 0 {
					hint = localizer.DGetf(domain, hint, args...)
				} else {
					hint = localizer.DGet(domain, hint)
				}
			} else {
				if len(args) > 0 {
					hint = localizer.Getf(hint, args...)
				} else {
					hint = localizer.Get(hint)
				}
			}
			return errors.WithHint(err, hint)
		}
		return err
	}
}
