package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse estructura estándar para todas las respuestas de la API
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// ErrorInfo información detallada del error
type ErrorInfo struct {
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// Success devuelve una respuesta exitosa
func Success(c *gin.Context, message string, data interface{}) {
	response := APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// Error devuelve una respuesta de error
func Error(c *gin.Context, statusCode int, message string) {
	response := APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Details: message,
		},
		Timestamp: time.Now(),
	}

	c.JSON(statusCode, response)
}

// ErrorWithCode devuelve una respuesta de error con código específico
func ErrorWithCode(c *gin.Context, statusCode int, code, message string) {
	response := APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    code,
			Details: message,
		},
		Timestamp: time.Now(),
	}

	c.JSON(statusCode, response)
}

// ValidationError devuelve un error de validación con detalles
func ValidationError(c *gin.Context, errors map[string]string) {
	response := APIResponse{
		Success: false,
		Message: "Errores de validación encontrados",
		Data:    gin.H{"validation_errors": errors},
		Error: &ErrorInfo{
			Code:    "VALIDATION_ERROR",
			Details: "Los datos proporcionados no pasaron la validación",
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusBadRequest, response)
}

// Created devuelve una respuesta de recurso creado
func Created(c *gin.Context, message string, data interface{}) {
	response := APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusCreated, response)
}

// NoContent devuelve una respuesta sin contenido
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Unauthorized devuelve una respuesta de no autorizado
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "No autorizado"
	}

	response := APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    "UNAUTHORIZED",
			Details: message,
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusUnauthorized, response)
}

// Forbidden devuelve una respuesta de prohibido
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "Acceso prohibido"
	}

	response := APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    "FORBIDDEN",
			Details: message,
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusForbidden, response)
}

// NotFound devuelve una respuesta de no encontrado
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "Recurso no encontrado"
	}

	response := APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    "NOT_FOUND",
			Details: message,
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusNotFound, response)
}

// InternalServerError devuelve una respuesta de error interno del servidor
func InternalServerError(c *gin.Context, message string) {
	if message == "" {
		message = "Error interno del servidor"
	}

	response := APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    "INTERNAL_SERVER_ERROR",
			Details: message,
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusInternalServerError, response)
}

// BadRequest devuelve una respuesta de petición incorrecta
func BadRequest(c *gin.Context, message string) {
	if message == "" {
		message = "Petición incorrecta"
	}

	response := APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    "BAD_REQUEST",
			Details: message,
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusBadRequest, response)
}
