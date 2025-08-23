package errors

import (
	"errors"
	"strings"
)

func (m mError) Meta() map[string]string {
	parts := strings.Split(m.meta, ";")
	meta := make(map[string]string)
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			continue
		}
		meta[kv[0]] = kv[1]
	}
	return meta
}

func Meta(err error) map[string]string {
	var e mError
	if errors.As(err, &e) {
		return e.Meta()
	}

	return nil
}

func WithMeta(err error, kv ...string) error {
	if len(kv)%2 != 0 {
		panic("WithMeta: odd number of key value pairs")
	}
	if err == nil {
		return err
	}

	metaMap := make(map[string]string)
	for i := 0; i < len(kv); i += 2 {
		metaMap[kv[i]] = kv[i+1]
	}

	parts := make([]string, 0, len(metaMap))
	for key, value := range metaMap {
		parts = append(parts, key+"="+value)
	}
	var e mError
	if errors.As(err, &e) {
		e.meta = strings.Join(parts, ";")
		return e
	}
	e = new(err.Error())
	e.meta = strings.Join(parts, ";")
	return e
}

func WithAddedMeta(err error, kv ...string) error {
	if len(kv)%2 != 0 {
		panic("WithAddedMeta: odd number of arguments")
	}
	if err == nil {
		return err
	}
	oldMeta := Meta(err)
	newKV := make([]string, 0, len(oldMeta)*2+len(kv))
	for k, v := range oldMeta {
		newKV = append(newKV, k, v)
	}
	newKV = append(newKV, kv...)

	return WithMeta(err, newKV...)
}
