package web

import (
	"fmt"
	"net/http"
)

var (
	// Text is a static singleton text result provider.
	Text TextResultProvider

	// assert TestResultProvider implements result provider.
	_ ResultProvider = Text
)

// TextResultProvider is the default response provider if none is specified.
type TextResultProvider struct{}

// NotFound returns a text response.
func (trp TextResultProvider) NotFound() Result {
	return &RawResult{
		StatusCode:  http.StatusNotFound,
		ContentType: ContentTypeText,
		Body:        []byte("Not Found"),
	}
}

// NotAuthorized returns a text response.
func (trp TextResultProvider) NotAuthorized() Result {
	return &RawResult{
		StatusCode:  http.StatusForbidden,
		ContentType: ContentTypeText,
		Body:        []byte("Not Authorized"),
	}
}

// InternalError returns a text response.
func (trp TextResultProvider) InternalError(err error) Result {
	return resultWithLoggedError(&RawResult{
		StatusCode:  http.StatusInternalServerError,
		ContentType: ContentTypeText,
		Body:        []byte(fmt.Sprintf("%+v", err)),
	}, err)
}

// BadRequest returns a text response.
func (trp TextResultProvider) BadRequest(err error) Result {
	if err != nil {
		return &RawResult{
			StatusCode:  http.StatusBadRequest,
			ContentType: ContentTypeText,
			Body:        []byte(fmt.Sprintf("Bad Request: %v", err)),
		}
	}
	return &RawResult{
		StatusCode:  http.StatusBadRequest,
		ContentType: ContentTypeText,
		Body:        []byte("Bad Request"),
	}
}

// Status returns a plaintext result.
func (trp TextResultProvider) Status(statusCode int, response ...interface{}) Result {
	return &RawResult{
		StatusCode:  statusCode,
		ContentType: ContentTypeText,
		Body:        []byte(fmt.Sprintf("%v", ResultOrDefault(http.StatusText(statusCode), response...))),
	}
}

// Result returns an xml response.
func (trp TextResultProvider) Result(result interface{}) Result {
	return &RawResult{
		StatusCode:  http.StatusOK,
		ContentType: ContentTypeText,
		Body:        []byte(fmt.Sprintf("%v", result)),
	}
}
