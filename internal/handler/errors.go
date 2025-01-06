package handler

import (
	"net/http"
)

type FieldErrMap map[string][]string

func (m *FieldErrMap) Empty() bool {
	return len(*m) == 0
}

func (m *FieldErrMap) AppendErr(key string, err error) {
	if *m == nil {
		*m = map[string][]string{}
	}

	if _, ok := (*m)[key]; !ok && err != nil {
		(*m)[key] = []string{}
	}

	if err != nil {
		(*m)[key] = append((*m)[key], err.Error())
	}
}

func (m *FieldErrMap) Append(key, value string) {
	if *m == nil {
		*m = map[string][]string{}
	}

	if _, ok := (*m)[key]; !ok {
		(*m)[key] = []string{}
	}

	(*m)[key] = append((*m)[key], value)
}

type validationError struct {
	Errors FieldErrMap `json:"errors"`
}

func NewValidationError(errors FieldErrMap) *validationError {
	return &validationError{
		Errors: errors,
	}
}

func (e *validationError) Write(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	JSON(w, e)
}

type Error struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func NewError(message string, statusCode int) *Error {
	return &Error{
		Message:    message,
		StatusCode: statusCode,
	}
}

func (e *Error) Write(w http.ResponseWriter) {
	w.WriteHeader(e.StatusCode)
	JSON(w, e)
}

func InternalServerError(w http.ResponseWriter) {
	NewError("internal server error", http.StatusInternalServerError).
		Write(w)
}

func InvalidJSON(w http.ResponseWriter) {
	NewError("invalid json", http.StatusBadRequest).
		Write(w)
}

func ValidationError(w http.ResponseWriter, errs FieldErrMap) {
	NewValidationError(errs).Write(w)
}

func NotFoundError(w http.ResponseWriter) {
	NewError("Resource not found", http.StatusNotFound).
		Write(w)
}

func UnauthorizedError(w http.ResponseWriter) {
	NewError("Unauthorized", http.StatusUnauthorized).
		Write(w)
}

func AlreayExistsError(w http.ResponseWriter) {
	NewError("resource already exists", http.StatusBadRequest).
		Write(w)
}
