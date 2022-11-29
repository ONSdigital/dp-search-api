package errors

import "errors"

// Error represents a handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type Error interface {
	error
	Status() int
}

// StatusError represents an error with an associated HTTP status code.
type StatusError struct {
	Code int
	Err  error
}

// Allows StatusError to satisfy the error interface.
func (e StatusError) Error() string {
	if e.Err == nil {
		return "nil"
	}

	return e.Err.Error()
}

// Status returns the HTTP status code.
func (e StatusError) Status() int {
	return e.Code
}

func ErrorStatus(err error) int {
	var rerr Error
	if errors.As(err, &rerr) {
		return rerr.Status()
	}

	return 0
}

func ErrorMessage(err error) string {
	var rerr Error
	if errors.As(err, &rerr) {
		if message := rerr.Error(); message != "" {
			return message
		}
	}

	return err.Error()
}
