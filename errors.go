package partupload

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// NewHttpError creates new HttpError instance.
func NewHttpError(code int, msg ...any) error {
	if code < 400 || code >= 600 {
		code = http.StatusInternalServerError
	}

	message := strings.TrimSpace(fmt.Sprint(msg...))
	if message == "" {
		message = http.StatusText(code)
	}

	return &HttpError{status: code, message: message}
}

// HttpErrorFromAny converts any error to the HttpError.
// Returns nil if no error is given.
func HttpErrorFromAny(err error) *HttpError {
	if err == nil {
		return nil
	}

	if errors.Is(err, &HttpError{}) {
		return err.(*HttpError)
	}

	err = NewHttpError(http.StatusInternalServerError, err.Error())

	return err.(*HttpError)
}

// HttpError is an error with an HTTP response status code.
type HttpError struct {
	status  int
	message string
}

// GetStatus returns an error's response status code.
func (e HttpError) GetStatus() int {
	return e.status
}

func (e HttpError) Error() string {
	return e.message
}
