package services

import (
	"client-data-compiler/internal/domain/errors"
	"client-data-compiler/internal/domain/models"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type ExcelService interface {
	ReadExcelFile(filePath string) ([]*models.Client, error)
	WriteExcelFile(clients []*models.Client, filePath string) error
	ValidateExcelStructure(filePath string) error
}

type excelService struct{}

func NewExcelService() ExcelService {
	return &excelService{}
}

// ReadExcelFile lee un archivo Excel y devuelve una lista de clientes
func (s *excelService) ReadExcelFile(filePath string) ([]*models.Client, error) {
	log.Printf("Iniciando lectura del archivo Excel: %s", filePath)

	// Verificar extensión del archivo
	if !strings.HasSuffix(strings.ToLower(filePath), ".xlsx") {
		log.Printf("Extensión de archivo inválida: %s", filePath)
		return nil, errors.ErrInvalidFileFormat
	}

	// Abrir archivo Excel
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Printf("Error abriendo archivo Excel %s: %v", filePath, err)
		return nil, errors.NewFileProcessingError(fmt.Sprintf("Error abriendo archivo: %v", err))
	}
	defer f.Close()

	log.Printf("Archivo Excel abierto exitosamente")

	// Obtener la primera hoja
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		log.Printf("No se encontraron hojas en el archivo")
		return nil, errors.NewFileProcessingError("No se encontraron hojas en el archivo")
	}

	log.Printf("Procesando hoja: %s", sheetName)

	// Obtener todas las filas
	rows, err := f.GetRows(sheetName)
	if err != nil {
		log.Printf("Error leyendo filas de la hoja %s: %v", sheetName, err)
		return nil, errors.NewFileProcessingError(fmt.Sprintf("Error leyendo filas: %v", err))
	}

	log.Printf("Total de filas encontradas: %d", len(rows))

	if len(rows) < 1 {
		log.Printf("Archivo sin contenido")
		return nil, errors.ErrFileEmpty
	}

	if len(rows) < 2 {
		log.Printf("Archivo solo tiene encabezados, sin datos")
		return nil, errors.NewFileProcessingError("El archivo solo contiene encabezados, sin datos")
	}

	// Validar estructura del encabezado
	log.Printf("Validando encabezados: %v", rows[0])
	if err := s.validateHeaders(rows[0]); err != nil {
		log.Printf("Error en validación de encabezados: %v", err)
		return nil, err
	}

	log.Printf("Encabezados validados correctamente")

	// Procesar datos
	var clients []*models.Client
	for i, row := range rows[1:] { // Saltar encabezado
		rowNumber := i + 2 // +2 porque empezamos en fila 1 y saltamos encabezado

		log.Printf("Procesando fila %d: %v", rowNumber, row)

		// Asegurar que la fila tenga al menos 4 columnas
		for len(row) < 4 {
			row = append(row, "")
		}

		client := &models.Client{
			ID:        i + 1,
			Clave:     strings.TrimSpace(row[0]),
			Nombre:    strings.TrimSpace(row[1]),
			Correo:    strings.TrimSpace(row[2]),
			Telefono:  strings.TrimSpace(row[3]),
			RowNumber: rowNumber,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Errors:    make(map[string]string),
			IsValid:   true,
		}

		// Log del cliente creado
		log.Printf("Cliente creado: ID=%d, Clave=%s, Nombre=%s, Correo=%s, Telefono=%s",
			client.ID, client.Clave, client.Nombre, client.Correo, client.Telefono)

		clients = append(clients, client)
	}

	log.Printf("Procesamiento completado: %d clientes creados", len(clients))
	return clients, nil
}

// WriteExcelFile escribe una lista de clientes a un archivo Excel
func (s *excelService) WriteExcelFile(clients []*models.Client, filePath string) error {
	log.Printf("Iniciando escritura de archivo Excel: %s con %d clientes", filePath, len(clients))

	f := excelize.NewFile()
	defer f.Close()

	// Crear hoja principal
	sheetName := "Clientes"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		log.Printf("Error creando hoja %s: %v", sheetName, err)
		return errors.NewFileProcessingError(fmt.Sprintf("Error creando hoja: %v", err))
	}

	// Establecer la hoja como activa
	f.SetActiveSheet(index)

	// Escribir encabezados
	headers := []string{"Clave", "Nombre", "Correo", "Telefono"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// Aplicar estilo a los encabezados
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E6E6FA"},
			Pattern: 1,
		},
	})
	f.SetCellStyle(sheetName, "A1", "D1", headerStyle)

	// Escribir datos
	for i, client := range clients {
		row := i + 2 // +2 porque empezamos en fila 2 (después del encabezado)

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), client.Clave)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), client.Nombre)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), client.Correo)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), client.Telefono)

		// Resaltar filas con errores
		if !client.IsValid {
			errorStyle, _ := f.NewStyle(&excelize.Style{
				Fill: excelize.Fill{
					Type:    "pattern",
					Color:   []string{"#FFE6E6"},
					Pattern: 1,
				},
			})
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), errorStyle)
		}
	}

	// Ajustar ancho de columnas
	f.SetColWidth(sheetName, "A", "A", 15) // Clave
	f.SetColWidth(sheetName, "B", "B", 30) // Nombre
	f.SetColWidth(sheetName, "C", "C", 35) // Correo
	f.SetColWidth(sheetName, "D", "D", 20) // Telefono

	// Crear hoja de errores si hay clientes inválidos
	hasErrors := false
	for _, client := range clients {
		if !client.IsValid {
			hasErrors = true
			break
		}
	}

	if hasErrors {
		s.createErrorSheet(f, clients)
	}

	// Guardar archivo
	if err := f.SaveAs(filePath); err != nil {
		log.Printf("Error guardando archivo %s: %v", filePath, err)
		return errors.NewFileProcessingError(fmt.Sprintf("Error guardando archivo: %v", err))
	}

	log.Printf("Archivo Excel guardado exitosamente: %s", filePath)
	return nil
}

// validateHeaders valida que los encabezados sean correctos
func (s *excelService) validateHeaders(headers []string) error {
	expectedHeaders := []string{"clave", "nombre", "correo", "telefono"}

	if len(headers) < 4 {
		return errors.NewFileProcessingError(
			"El archivo debe tener al menos 4 columnas: Clave, Nombre, Correo, Telefono",
		)
	}

	// Normalizar encabezados (minúsculas, sin espacios)
	for i, header := range headers[:4] {
		normalized := strings.ToLower(strings.TrimSpace(header))
		normalized = strings.ReplaceAll(normalized, " ", "")
		normalized = strings.ReplaceAll(normalized, "é", "e")
		normalized = strings.ReplaceAll(normalized, "teléfono", "telefono")

		log.Printf("Comparando encabezado %d: '%s' (normalizado: '%s') con esperado: '%s'",
			i+1, header, normalized, expectedHeaders[i])

		if normalized != expectedHeaders[i] {
			return errors.NewFileProcessingError(
				fmt.Sprintf("Encabezado incorrecto en columna %d. Se esperaba '%s', se encontró '%s'",
					i+1, expectedHeaders[i], header),
			)
		}
	}

	return nil
}

// ValidateExcelStructure valida que el archivo Excel tenga la estructura correcta
func (s *excelService) ValidateExcelStructure(filePath string) error {
	log.Printf("Validando estructura del archivo Excel: %s", filePath)

	// Verificar extensión
	if !strings.HasSuffix(strings.ToLower(filePath), ".xlsx") {
		return errors.ErrInvalidFileFormat
	}

	// Abrir archivo
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Printf("Error abriendo archivo para validación: %v", err)
		return errors.NewFileProcessingError(fmt.Sprintf("Error abriendo archivo: %v", err))
	}
	defer f.Close()

	// Verificar que tenga al menos una hoja
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return errors.ErrInvalidExcelStructure
	}

	// Obtener filas
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return errors.NewFileProcessingError(fmt.Sprintf("Error leyendo archivo: %v", err))
	}

	if len(rows) < 1 {
		return errors.ErrFileEmpty
	}

	// Validar encabezados
	return s.validateHeaders(rows[0])
}

// createErrorSheet crea una hoja con el detalle de errores
func (s *excelService) createErrorSheet(f *excelize.File, clients []*models.Client) {
	errorSheetName := "Errores"
	f.NewSheet(errorSheetName)

	// Encabezados de la hoja de errores
	errorHeaders := []string{"Fila", "Clave", "Nombre", "Campo", "Error"}
	for i, header := range errorHeaders {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(errorSheetName, cell, header)
	}

	// Estilo para encabezados de errores
	errorHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFE6E6"},
			Pattern: 1,
		},
	})
	f.SetCellStyle(errorSheetName, "A1", "E1", errorHeaderStyle)

	// Escribir errores
	errorRow := 2
	for _, client := range clients {
		if !client.IsValid {
			for field, errorMsg := range client.Errors {
				f.SetCellValue(errorSheetName, fmt.Sprintf("A%d", errorRow), client.RowNumber)
				f.SetCellValue(errorSheetName, fmt.Sprintf("B%d", errorRow), client.Clave)
				f.SetCellValue(errorSheetName, fmt.Sprintf("C%d", errorRow), client.Nombre)
				f.SetCellValue(errorSheetName, fmt.Sprintf("D%d", errorRow), field)
				f.SetCellValue(errorSheetName, fmt.Sprintf("E%d", errorRow), errorMsg)
				errorRow++
			}
		}
	}

	// Ajustar ancho de columnas en hoja de errores
	f.SetColWidth(errorSheetName, "A", "A", 8)  // Fila
	f.SetColWidth(errorSheetName, "B", "B", 15) // Clave
	f.SetColWidth(errorSheetName, "C", "C", 25) // Nombre
	f.SetColWidth(errorSheetName, "D", "D", 15) // Campo
	f.SetColWidth(errorSheetName, "E", "E", 50) // Error
}
