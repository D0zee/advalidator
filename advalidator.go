package advalidator

import (
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

var ErrNotStruct = errors.New("wrong argument given, should be a struct")
var ErrInvalidValidatorSyntax = errors.New("invalid validator syntax")
var ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")
var ErrValidatorLen = errors.New("validator LEN: wrong length of string")

type ValidationError struct {
	Err error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var s string
	for _, err := range v {
		s += err.Err.Error()
	}
	return s
}

type Field struct {
	Value    reflect.Value
	TagValue string
}

func getValueOfTag(f Field, prefix string) (string, error) {
	tagValue := f.TagValue
	after := strings.TrimPrefix(tagValue, prefix)
	if after == "" {
		return "", ErrInvalidValidatorSyntax
	}
	return after, nil
}

func validateLen(errs ValidationErrors, f Field) ValidationErrors {
	after, err := getValueOfTag(f, "len:")
	if err != nil {
		return append(errs, ValidationError{err})
	}

	bound, err := strconv.Atoi(after)
	if err != nil || bound < 1 {
		return append(errs, ValidationError{ErrInvalidValidatorSyntax})
	}
	if f.Value.Kind() == reflect.String {
		if !(len(f.Value.String()) > 0 && len(f.Value.String()) < bound) {
			return append(errs, ValidationError{ErrValidatorLen})
		}
	}
	return errs
}

// Validate that string is less than int in tagValue and that string is nonempty
func Validate(val any) error {
	typeV := reflect.TypeOf(val)
	valueV := reflect.ValueOf(val)
	if typeV.Kind() != reflect.Struct {
		return ErrNotStruct
	}
	n := valueV.NumField()
	errs := ValidationErrors{}
	for i := 0; i < n; i++ {
		ft := typeV.Field(i)
		vt := valueV.Field(i)
		tv := ft.Tag.Get("validate")
		if tv == "" {
			continue
		}
		if !ft.IsExported() {
			errs = append(errs, ValidationError{ErrValidateForUnexportedFields})
			continue
		}
		if strings.HasPrefix(tv, "len:") {
			errs = validateLen(errs, Field{Value: vt, TagValue: tv})
		} else {
			errs = append(errs, ValidationError{ErrInvalidValidatorSyntax})
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}
