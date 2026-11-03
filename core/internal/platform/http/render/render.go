package httprender

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func Empty(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Error(w http.ResponseWriter, statusCode int, err error) {
	JSON(w, statusCode, &ErrorResponse{Code: statusCode, Message: err.Error()})
}

func InvalidJSON(w http.ResponseWriter) {
	Error(w, http.StatusBadRequest, errors.New("invalid json"))
}

func NotFound(w http.ResponseWriter, err error) {
	Error(w, http.StatusNotFound, err)
}

func Conflict(w http.ResponseWriter, err error) {
	Error(w, http.StatusConflict, err)
}

// Unauthorized answers without WWW-Authenticate on purpose: the browser must not
// show its own Basic dialog, the SPA has a login form.
func Unauthorized(w http.ResponseWriter, err error) {
	Error(w, http.StatusUnauthorized, err)
}

func ValidationError(w http.ResponseWriter, err error, request any) {
	result, ok := errors.AsType[validator.ValidationErrors](err)
	if !ok || len(result) == 0 {
		Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	name := jsonTag(request, result[0].Field())
	if name == "" {
		name = result[0].Field()
	}
	Error(w, http.StatusUnprocessableEntity, fmt.Errorf("invalid field: %s", name))
}

// Internal hides the cause from the client and logs it with the route, because
// nothing else in the stack will.
func Internal(w http.ResponseWriter, r *http.Request, err error) {
	slog.ErrorContext(r.Context(), "request failed",
		slog.String("method", r.Method), slog.String("path", r.URL.Path), slog.String("error", err.Error()))
	Error(w, http.StatusInternalServerError, errors.New("internal server error"))
}

func jsonTag(s any, fieldName string) string {
	t := reflect.TypeOf(s)
	if t == nil {
		return ""
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return ""
	}
	field, found := t.FieldByName(fieldName)
	if !found {
		return ""
	}
	return field.Tag.Get("json")
}
