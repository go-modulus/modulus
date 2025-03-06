package validator

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-modulus/modulus/errors/erruser"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"golang.org/x/text/message"
)

type ErrInvalidInput struct {
	Fields []InvalidField
}

type InvalidField struct {
	Name    string
	Message string
	Code    string
}

func (e ErrInvalidInput) Code() string {
	return "InvalidInput"
}

func (e ErrInvalidInput) Message(p *message.Printer) string {
	return p.Sprintf("Invalid input")
}

func (e ErrInvalidInput) Details() map[string]any {
	fields := make(map[string]map[string]string)
	for _, field := range e.Fields {
		fields[field.Name] = map[string]string{
			"code":    field.Code,
			"message": field.Message,
		}
	}

	return map[string]any{
		"fields": fields,
	}
}

func (e ErrInvalidInput) Error() string {
	var fields []string
	for _, field := range e.Fields {
		fields = append(fields, fmt.Sprintf("%s: %s: %s", field.Name, field.Code, field.Message))
	}
	return fmt.Sprintf("invalid input (%s)", strings.Join(fields, ", "))
}

//func AsOzzoError(ctx context.Context, err error) validation.Error {
//	p := translationContext.GetPrinter(ctx)
//	return validation.NewError(
//		err.Error(),
//		errors2.Message(p, err),
//	)
//}

func ConvertOzzoError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	var fieldErrors validation.Errors
	if !errors.As(err, &fieldErrors) {
		return err
	}

	path := graphql.GetPath(ctx)
	fields := goFieldsRecursive(err, path.String())

	if len(fields) == 0 {
		return err
	}

	errs := make([]error, 0, len(fields))
	for _, field := range fields {
		errs = append(errs, erruser.New(field.Name+"."+field.Code, field.Message))
	}

	return erruser.NewValidationError(errs...)
}

func goFieldsRecursive(err error, path string) []InvalidField {
	fields := make([]InvalidField, 0)
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
			field := path + "." + strings.ToLower(key)
			fields = append(fields, NewInvalidFieldFromOzzo(field, innerErr))
		}
		innerErrs, ok2 := fieldErr.(validation.Errors)
		if ok2 {
			fields = append(fields, goFieldsRecursive(innerErrs, path+"."+key)...)
		}
	}
	return fields
}

func NewInvalidFieldFromOzzo(field string, err validation.Error) InvalidField {
	return InvalidField{
		Name:    field,
		Code:    strings.Replace(err.Code(), "_", ".", 1),
		Message: err.Error(),
	}
}

func Path(ctx context.Context, path ...string) string {
	path = append([]string{graphql.GetPath(ctx).String()}, path...)
	return strings.Join(path, ".")
}

type Validatable interface {
	Validate(ctx context.Context) error
}

// ValidateStructWithContext it is a wrapper around ozzo-validation ValidateStructWithContext
func ValidateStructWithContext(ctx context.Context, structPtr interface{}, fields ...*validation.FieldRules) error {
	err := validation.ValidateStructWithContext(ctx, structPtr, fields...)
	if err != nil {
		return ConvertOzzoError(ctx, err)
	}
	return nil
}
