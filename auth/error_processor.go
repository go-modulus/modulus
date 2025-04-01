package auth

import (
	"context"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/http/errhttp"
	"net/http"
	"slices"
)

func AddHttpCode() errhttp.ErrorProcessor {
	return func(ctx context.Context, err error) error {
		if err == nil {
			return nil
		}
		tags := errors.Tags(err)
		if slices.Contains(tags, TagUnauthenticated) {
			err = errhttp.ErrWithHttpCode(err, http.StatusUnauthorized)
		} else if slices.Contains(tags, TagUnauthorized) {
			err = errhttp.ErrWithHttpCode(err, http.StatusForbidden)
		}
		return err
	}
}
