package errors

import "errors"

type withMeta struct {
	meta map[string]string
	err  error
}

func (m withMeta) Meta() map[string]string {
	return m.meta
}

func (m withMeta) Error() string {
	return m.err.Error()
}

func (m withMeta) Unwrap() error {
	return m.err
}

func Meta(err error) map[string]string {
	type withMeta interface {
		Meta() map[string]string
	}
	var we withMeta
	if errors.As(err, &we) {
		return we.Meta()
	}
	return nil
}

func WrapMeta(err error, kv ...string) error {
	if err == nil {
		return err
	}
	meta := make(map[string]string)
	for i := 0; i < len(kv); i += 2 {
		meta[kv[i]] = kv[i+1]
	}
	return withMeta{meta: meta, err: err}
}
