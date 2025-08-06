// Package http provides HTTP response utilities.
// This is part of the Adapters layer in Hexagonal Architecture.
package utils

import (
	"encoding/json"
	"net/http"
)

// APIResponse represents the standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ResponseWriter provides utility methods for writing standardized HTTP responses
type ResponseWriter struct{}

// NewResponseWriter creates a new ResponseWriter instance
func NewResponseWriter() *ResponseWriter {
	return &ResponseWriter{}
}

// WriteSuccess writes a successful response with data
func (rw *ResponseWriter) WriteSuccess(w http.ResponseWriter, message string, data interface{}) {
	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	rw.writeJSON(w, http.StatusOK, response)
}

// WriteCreated writes a successful creation response
func (rw *ResponseWriter) WriteCreated(w http.ResponseWriter, message string, data interface{}) {
	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	rw.writeJSON(w, http.StatusCreated, response)
}

// WriteError writes an error response
func (rw *ResponseWriter) WriteError(w http.ResponseWriter, statusCode int, message string) {
	response := APIResponse{
		Success: false,
		Message: message,
	}
	rw.writeJSON(w, statusCode, response)
}

// WriteBadRequest writes a bad request error response
func (rw *ResponseWriter) WriteBadRequest(w http.ResponseWriter, message string) {
	rw.WriteError(w, http.StatusBadRequest, message)
}

// WriteUnauthorized writes an unauthorized error response
func (rw *ResponseWriter) WriteUnauthorized(w http.ResponseWriter, message string) {
	rw.WriteError(w, http.StatusUnauthorized, message)
}

// WriteForbidden writes a forbidden error response
func (rw *ResponseWriter) WriteForbidden(w http.ResponseWriter, message string) {
	rw.WriteError(w, http.StatusForbidden, message)
}

// WriteNotFound writes a not found error response
func (rw *ResponseWriter) WriteNotFound(w http.ResponseWriter, message string) {
	rw.WriteError(w, http.StatusNotFound, message)
}

// WriteInternalServerError writes an internal server error response
func (rw *ResponseWriter) WriteInternalServerError(w http.ResponseWriter, message string) {
	rw.WriteError(w, http.StatusInternalServerError, message)
}

// writeJSON writes a JSON response with the specified status code
func (rw *ResponseWriter) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, write a simple error response
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success":false,"message":"Internal server error"}`)) 
	}
}

// Global response writer instance for convenience
var Response = NewResponseWriter()