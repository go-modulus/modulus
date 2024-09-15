package errhttp

import (
	"errors"
	"fmt"
	"net/http"
)

type withStatus struct {
	err    error
	status int
}

func (w withStatus) Status() int {
	return w.status
}

func (w withStatus) Error() string {
	return fmt.Sprintf("status=%d %s", w.status, w.err)
}

func (w withStatus) Unwrap() error { return w.err }

func Wrap(err error, status int) error {
	return withStatus{
		err:    err,
		status: status,
	}
}

func With(status int) func(error) error {
	return func(err error) error {
		return Wrap(err, status)
	}
}

func Status(err error) int {
	type withStatus interface {
		Status() int
	}
	var wc withStatus
	if errors.As(err, &wc) {
		return wc.Status()
	}
	return http.StatusInternalServerError
}
