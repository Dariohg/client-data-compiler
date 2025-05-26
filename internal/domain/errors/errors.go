package errors

import "fmt"

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

// Errores comunes
var (
	ErrInvalidFileFormat = &AppError{
		Code:    "INVALID_FILE_FORMAT",
		Message: "Formato de archivo inválido. Solo se permiten archivos Excel (.xlsx)",
	}

	ErrFileEmpty = &AppError{
		Code:    "FILE_EMPTY",
		Message: "El archivo está vacío",
	}

	ErrClientNotFound = &AppError{
		Code:    "CLIENT_NOT_FOUND",
		Message: "Cliente no encontrado",
	}

	ErrInvalidClientID = &AppError{
		Code:    "INVALID_CLIENT_ID",
		Message: "ID de cliente inválido",
	}

	ErrDuplicateClientKey = &AppError{
		Code:    "DUPLICATE_CLIENT_KEY",
		Message: "Ya existe un cliente con esta clave",
	}

	ErrInvalidExcelStructure = &AppError{
		Code:    "INVALID_EXCEL_STRUCTURE",
		Message: "La estructura del archivo Excel no es válida",
	}
)

// Funciones para crear errores específicos
func NewValidationError(field, message string) *AppError {
	return &AppError{
		Code:    "VALIDATION_ERROR",
		Message: fmt.Sprintf("Error de validación en %s: %s", field, message),
	}
}

func NewFileProcessingError(message string) *AppError {
	return &AppError{
		Code:    "FILE_PROCESSING_ERROR",
		Message: fmt.Sprintf("Error procesando archivo: %s", message),
	}
}

func NewDatabaseError(message string) *AppError {
	return &AppError{
		Code:    "DATABASE_ERROR",
		Message: fmt.Sprintf("Error en base de datos: %s", message),
	}
}
