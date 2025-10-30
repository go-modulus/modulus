package validator

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-modulus/modulus/errors/erruser"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type invalidField struct {
	Name    string
	Message string
	Code    string
}

func convertOzzoError(err error, structName string) error {
	if err == nil {
		return nil
	}

	var fieldErrors validation.Errors
	if !errors.As(err, &fieldErrors) {
		return err
	}

	fields := goFieldsRecursive(err, structName)

	if len(fields) == 0 {
		return err
	}

	errs := make([]error, 0, len(fields))
	for _, field := range fields {
		errs = append(errs, erruser.New(field.Name, field.Message))
	}

	return erruser.NewValidationError(errs...)
}

func makeField(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + strings.ToLower(key)
}

func goFieldsRecursive(err error, path string) []invalidField {
	fields := make([]invalidField, 0)
	if err == nil {
		return fields
	}
	var fieldErrors validation.Errors
	if !errors.As(err, &fieldErrors) {
		return fields
	}
	for key, fieldErr := range fieldErrors {
		innerErr, ok := fieldErr.(validation.ErrorObject)
		if ok {
			field := makeField(path, key)
			fields = append(fields, newInvalidFieldFromOzzo(field, innerErr))
		}
		innerErrs, ok2 := fieldErr.(validation.Errors)
		if ok2 {
			fields = append(fields, goFieldsRecursive(innerErrs, makeField(path, key))...)
		}
	}
	return fields
}

func newInvalidFieldFromOzzo(field string, err validation.Error) invalidField {
	return invalidField{
		Name:    field,
		Code:    strings.Replace(err.Code(), "_", ".", 1),
		Message: err.Error(),
	}
}

type Validatable interface {
	Validate(ctx context.Context) error
}

// ValidateStructWithContext it is a wrapper around ozzo-validation ValidateStructWithContext
func ValidateStructWithContext(ctx context.Context, structPtr interface{}, fields ...*validation.FieldRules) error {
	err := validation.ValidateStructWithContext(ctx, structPtr, fields...)
	if err != nil {
		structName := strings.Split(fmt.Sprintf("%T", structPtr), ".")[1]
		return convertOzzoError(err, structName)
	}
	return nil
}

func ValidateWithContext(ctx context.Context, valuePtr interface{}, valueName string, rules ...validation.Rule) error {
	err := validation.ValidateWithContext(
		ctx,
		valuePtr,
		rules...,
	)
	if err != nil {
		return convertOzzoError(err, valueName)
	}
	return nil
}
