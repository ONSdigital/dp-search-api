package query

import (
	"context"
	"errors"
	"fmt"
	"strconv"
)

// ParamValidator is a map of parameters to their corresponding validator func
type ParamValidator map[paramName]validator

// Validate calls the validator func for the provided parameter name and value
func (qpv ParamValidator) Validate(_ context.Context, name, value string) (interface{}, error) {
	if v, ok := qpv[paramName(name)]; ok {
		return v(value)
	}
	return nil, fmt.Errorf("cannot validate: no validator for %s", name)
}

type validator func(param string) (interface{}, error)
type paramName string

// NewReleaseQueryParamValidator creates a validator to validate
// parameters for the Release endpoint
func NewReleaseQueryParamValidator() ParamValidator {
	return ParamValidator{
		"limit":        validateLimit,
		"offset":       validateOffset,
		"date":         validateDate,
		"sort":         validateSort,
		"release-type": validateReleaseType,
	}
}

// NewSearchQueryParamValidator creates a validator to validate
// parameters for the Search endpoint
func NewSearchQueryParamValidator() ParamValidator {
	return ParamValidator{
		"limit":  validateLimit,
		"offset": validateOffset,
		"sort": func(param string) (interface{}, error) {
			return param, nil
		},
	}
}

var validateLimit validator = func(param string) (interface{}, error) {
	value, err := strconv.Atoi(param)
	if err != nil {
		return 0, errors.New("limit search parameter provided with non numeric characters")
	}
	if value < 0 {
		return 0, errors.New("limit search parameter provided with negative value")
	}
	if value > 1000 {
		return 0, errors.New("limit search parameter provided with a value that is too high")
	}

	return value, nil
}

var validateOffset validator = func(param string) (interface{}, error) {
	value, err := strconv.Atoi(param)
	if err != nil {
		return 0, errors.New("offset search parameter provided with non numeric characters")
	}
	if value < 0 {
		return 0, errors.New("offset search parameter provided with negative value")
	}
	return value, nil
}

var validateDate validator = func(param string) (interface{}, error) {
	value, err := ParseDate(param)
	if err != nil {
		return nil, fmt.Errorf("date search parameter provided is invalid: %w", err)
	}
	return value, nil
}

var validateSort validator = func(param string) (interface{}, error) {
	value, err := ParseSort(param)
	if err != nil {
		return nil, fmt.Errorf("sort search parameter provided is invalid: %w", err)
	}
	return value, nil
}

var validateReleaseType validator = func(param string) (interface{}, error) {
	value, err := ParseReleaseType(param)
	if err != nil {
		return nil, fmt.Errorf("release-type parameter provided is invalid: %w", err)
	}
	return value, nil
}
