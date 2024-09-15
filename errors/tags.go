package errors

import (
	"errors"
)

type withTags struct {
	tags []string
	err  error
}

func (m withTags) Tags() []string {
	return m.tags
}

func (m withTags) Error() string {
	return m.err.Error()
}

func (m withTags) Unwrap() error {
	return m.err
}

func Tags(err error) []string {
	if err == nil {
		return nil
	}
	type withTags interface {
		Tags() []string
	}
	var we withTags
	if errors.As(err, &we) {
		return we.Tags()
	}
	return nil
}

func WrapAddingTags(err error, tags ...string) error {
	if err == nil {
		return err
	}
	oldTags := Tags(err)
	tags = append(oldTags, tags...)

	return withTags{tags: tags, err: err}
}
