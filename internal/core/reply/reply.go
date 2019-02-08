// Package reply - common replies provided by server
package reply

import (
	"encoding/json"
	"net/http"

	"github.com/wtask/pwsrv/internal/api"
)

// ServiceUnavailable - returns http-handler to make "503. Service unavailable" error response.
func ServiceUnavailable() http.HandlerFunc {
	return jsonContent(http.StatusServiceUnavailable, &api.ErrorResponse{true, "Service unavailable"})
}

// BadRequest - returns http-handler to make bad request (400) response with custom error messsage.
func BadRequest(msg string) http.HandlerFunc {
	return jsonContent(http.StatusBadRequest, &api.ErrorResponse{true, msg})
}

// Unauthorized - returns http-handler to make "401. Unauthorized" error response.
func Unauthorized() http.HandlerFunc {
	return jsonContent(http.StatusUnauthorized, &api.ErrorResponse{true, "Unauthorized"})
}

// Forbidden - returns http-handler to make forbidden (403) error response.
func Forbidden(msg string) http.HandlerFunc {
	return jsonContent(http.StatusForbidden, &api.ErrorResponse{true, msg})
}

// Conflict - returns http-handler to make conflict (409) response with custom error message.
func Conflict(msg string) http.HandlerFunc {
	return jsonContent(http.StatusConflict, &api.ErrorResponse{true, msg})
}

// InternalServerError - returns http-handler to make server error (500) response with custom error message.
func InternalServerError(msg string) http.HandlerFunc {
	return jsonContent(http.StatusInternalServerError, &api.ErrorResponse{true, msg})
}

// OK - returns a handler to reply the successful (200) request processing.
func OK(data interface{}) http.HandlerFunc {
	return jsonContent(http.StatusOK, data)
}

// jsonContent - returns http handler to respond any given data with specified http-status code.
func jsonContent(status int, data interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(&data)
	}
}
