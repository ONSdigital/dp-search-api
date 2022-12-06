package sdk

import "errors"

var (

	// ErrGetPermissionsResponseBodyNil error used when a nil response is returned from the permissions API.
	ErrGetPermissionsResponseBodyNil = errors.New("error creating get permissions request http.Request required but was nil")
)
