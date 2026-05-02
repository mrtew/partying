package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func JSON(w http.ResponseWriter, statusCode int, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Success: success,
		Message: message,
		Data:    data,
	})
}

func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, true, "", data)
}

func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, true, "", data)
}

func BadRequest(w http.ResponseWriter, message string) {
	JSON(w, http.StatusBadRequest, false, message, nil)
}

func Unauthorized(w http.ResponseWriter, message string) {
	JSON(w, http.StatusUnauthorized, false, message, nil)
}

// func Forbidden(w http.ResponseWriter, message string) {
// 	JSON(w, http.StatusForbidden, false, message, nil)
// }

func InternalError(w http.ResponseWriter, message string) {
	JSON(w, http.StatusInternalServerError, false, message, nil)
}
