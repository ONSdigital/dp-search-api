package apierrors

import "errors"

// A list of error messages for Search API

var (
	ErrUnmarshallingJSON = errors.New("failed to parse json body")
	ErrMarshallingQuery  = errors.New("failed to marshal query to bytes for request body to send to elastic")
)
