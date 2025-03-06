package http

import (
	"braces.dev/errtrace"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ggicci/httpin"
	httpinCore "github.com/ggicci/httpin/core"
	"github.com/go-modulus/modulus/errors/erruser"
	"github.com/go-modulus/modulus/http/errhttp"
	translationContext "github.com/go-modulus/modulus/translation"
	"github.com/go-modulus/modulus/validator"
	"io"
	"net/http"
)

func ReadBody(req *http.Request) ([]byte, error) {
	body, err := io.ReadAll(req.Body)
	if err == nil {
		req.Body = io.NopCloser(bytes.NewReader(body))
	}
	return errtrace.Wrap2(body, err)
}

type RequestWithInput[I any] struct {
	req   *http.Request
	Input I
}

func (ri RequestWithInput[I]) Context() context.Context {
	return ri.req.Context()
}

func (ri RequestWithInput[I]) Req() *http.Request {
	return ri.req
}

func (ri RequestWithInput[I]) RawBody() ([]byte, error) {
	return ReadBody(ri.req)
}

type InputHandler[B any] func(w http.ResponseWriter, req RequestWithInput[B]) error

func WrapInputHandler[B any](handle InputHandler[B]) errhttp.Handler {
	var input B
	engine, err := httpin.New(input)
	if err != nil {
		panic(fmt.Errorf("modulus/http: %w", err))
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		body, err := ReadBody(r)
		if err != nil {
			return errtrace.Wrap(err)
		}

		input, err := engine.Decode(r)
		if err != nil {
			var fErr *httpinCore.InvalidFieldError
			if errors.As(err, &fErr) {
				t := translationContext.GetPrinter(r.Context())
				msg := t.Sprintf("Invalid value")
				switch fErr.Directive {
				case "required":
					msg = t.Sprintf("Required")
				case "nonzero":
					msg = t.Sprintf("Required to be non-zero value")
				}
				return erruser.NewValidationError(erruser.New(fErr.Key, msg))
			}
			return errtrace.Wrap(err)
		}

		r.Body = io.NopCloser(bytes.NewReader(body))

		if validatable, ok := input.(validator.Validatable); ok {
			err := validatable.Validate(r.Context())
			if err != nil {
				return errtrace.Wrap(err)
			}
		}
		typedInput, ok := input.(*B)
		if !ok {
			return errtrace.Errorf("invalid typed input %T != %T", input, typedInput)
		}
		return handle(w, RequestWithInput[B]{req: r, Input: *typedInput})
	}
}

type OptionalJsonDecoder struct{}

func (o *OptionalJsonDecoder) Decode(src io.Reader, dst interface{}) error {
	err := json.NewDecoder(src).Decode(dst)

	return err
}

type JSONBody struct{}

func (b *JSONBody) Decode(src io.Reader, dst any) error {
	err := json.NewDecoder(src).Decode(dst)
	if errors.Is(err, io.EOF) {
		err = nil
	}
	return err
}

func (b *JSONBody) Encode(src any) (io.Reader, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(src); err != nil {
		return nil, err
	}
	return &buf, nil
}

func init() {
	httpinCore.RegisterBodyFormat("optionalJson", &JSONBody{})
}
