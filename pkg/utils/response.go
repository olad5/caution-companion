package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	appErrors "github.com/olad5/go-hackathon-starter-template/pkg/errors"

	"go.uber.org/zap"
)

func SuccessResponse(w http.ResponseWriter, message string, data interface{}, l *zap.Logger) {
	type SuccessResponse struct {
		Status  bool        `json:"status"`
		Message string      `json:"message"`
		Data    interface{} `json:"data,omitempty"`
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(SuccessResponse{
		Status:  true,
		Message: message,
		Data:    data,
	}); err != nil {
		l.Error("Error sending response", zap.Error(err))
	}
}

func ErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	type SuccessResponse struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
	}
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(SuccessResponse{Status: false, Message: message}); err != nil {
		log.Printf("Error sending response: %v", err)
	}
}

func InternalServerErrorResponse(w http.ResponseWriter, err error, l *zap.Logger) {
	l.Error("[INTERNAL_SERVER_ERR]", zap.Error(err))
	ErrorResponse(w, appErrors.ErrSomethingWentWrong, http.StatusInternalServerError)
}

// TODO:TODO: I should use this Encode method
func Encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func Decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}
