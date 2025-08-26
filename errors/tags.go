package errors

import (
	"errors"
	"strings"
)

func (m mError) Tags() []string {
	tags := strings.Split(m.tags, ",")
	unique := make(map[string]struct{})

	var result []string
	for _, tag := range tags {
		if _, ok := unique[tag]; ok {
			continue
		}
		result = append(result, tag)
		unique[tag] = struct{}{}
	}
	return result
}
func (m mError) HasTag(tag string) bool {
	for _, t := range m.Tags() {
		if t == tag {
			return true
		}
	}
	return false
}

func Tags(err error) []string {
	if err == nil {
		return nil
	}
	var e mError
	if errors.As(err, &e) {
		return e.Tags()
	}
	return nil
}

func HasTag(err error, tag string) bool {
	if err == nil {
		return false
	}
	var e mError
	if errors.As(err, &e) {
		return e.HasTag(tag)
	}
	return false
}

func WithAddedTags(err error, tags ...string) error {
	if err == nil {
		return err
	}
	oldTags := Tags(err)
	tags = append(oldTags, tags...)

	e := copyErr(err)
	e.tags = strings.Join(tags, ",")
	return e

}
