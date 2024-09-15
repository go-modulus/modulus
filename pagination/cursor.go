package pagination

import (
	"braces.dev/errtrace"
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/go-modulus/modulus/erruser"
)

var ErrInvalidCursor = erruser.New("InvalidCursor", "Invalid cursor")

func DecodeCursor[T any](ctx context.Context, rawCursor string) (T, error) {
	var cursor T
	rawJson, err := base64.StdEncoding.DecodeString(rawCursor)
	if err != nil {
		return cursor, ErrInvalidCursor
	}

	err = json.Unmarshal(rawJson, &cursor)
	if err != nil {
		return cursor, ErrInvalidCursor
	}

	return cursor, nil
}

func EncodeCursor[T any](cursor T) (string, error) {
	rawJson, err := json.Marshal(cursor)
	if err != nil {
		return "", errtrace.Wrap(err)
	}
	return base64.StdEncoding.EncodeToString(rawJson), nil
}
