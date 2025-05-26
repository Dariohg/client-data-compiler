package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExcelColumnIndex convierte una letra de columna a índice numérico
func ExcelColumnIndex(col string) int {
	col = strings.ToUpper(col)
	result := 0
	for i, r := range col {
		result = result*26 + int(r-'A') + 1
		if i > 0 {
			result--
		}
	}
	return result - 1
}

// IndexToExcelColumn convierte un índice numérico a letra de columna
func IndexToExcelColumn(index int) string {
	result := ""
	for index >= 0 {
		result = string(rune('A'+index%26)) + result
		index = index/26 - 1
	}
	return result
}

// GetCellAddress obtiene la dirección de celda (ej: A1, B2)
func GetCellAddress(col, row int) string {
	return fmt.Sprintf("%s%d", IndexToExcelColumn(col), row+1)
}

// ValidateExcelFile verifica si un archivo es un Excel válido
func ValidateExcelFile(filePath string) error {
	// Verificar extensión
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".xlsx" && ext != ".xls" {
		return fmt.Errorf("archivo debe ser Excel (.xlsx o .xls), recibido: %s", ext)
	}

	// Verificar que el archivo existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("archivo no existe: %s", filePath)
	}

	// Intentar abrir el archivo
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("error abriendo archivo Excel: %v", err)
	}
	defer f.Close()

	return nil
}

// GetExcelFileInfo obtiene información básica del archivo Excel
func GetExcelFileInfo(filePath string) (map[string]interface{}, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Obtener información del archivo
	stat, _ := os.Stat(filePath)
	sheetList := f.GetSheetList()

	info := map[string]interface{}{
		"filename":    filepath.Base(filePath),
		"file_size":   stat.Size(),
		"modified":    stat.ModTime(),
		"sheets":      sheetList,
		"sheet_count": len(sheetList),
	}

	// Información de la primera hoja
	if len(sheetList) > 0 {
		rows, err := f.GetRows(sheetList[0])
		if err == nil {
			info["total_rows"] = len(rows)
			if len(rows) > 0 {
				info["total_columns"] = len(rows[0])
				info["data_rows"] = len(rows) - 1 // Excluyendo encabezado
			}
		}
	}

	return info, nil
}

// CleanExcelCell limpia el contenido de una celda Excel
func CleanExcelCell(value string) string {
	// Eliminar espacios extra
	value = strings.TrimSpace(value)

	// Reemplazar múltiples espacios por uno solo
	value = strings.Join(strings.Fields(value), " ")

	// Eliminar caracteres de control
	value = strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, value)

	return value
}

// ParseExcelDate parsea una fecha de Excel
func ParseExcelDate(value string) (time.Time, error) {
	// Formatos comunes de fecha en Excel
	formats := []string{
		"2006-01-02",
		"02/01/2006",
		"01/02/2006",
		"2006-01-02 15:04:05",
		"02/01/2006 15:04:05",
		"01-02-2006",
		"2006/01/02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("formato de fecha no válido: %s", value)
}

// CreateExcelStyle crea un estilo para Excel
func CreateExcelStyle(f *excelize.File, options map[string]interface{}) (int, error) {
	style := &excelize.Style{}

	// Font
	if font, ok := options["font"].(map[string]interface{}); ok {
		style.Font = &excelize.Font{}
		if bold, ok := font["bold"].(bool); ok {
			style.Font.Bold = bold
		}
		if size, ok := font["size"].(float64); ok {
			style.Font.Size = size
		}
		if color, ok := font["color"].(string); ok {
			style.Font.Color = color
		}
	}

	// Fill (background color)
	if fill, ok := options["fill"].(map[string]interface{}); ok {
		style.Fill = excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
		}
		if colors, ok := fill["color"].([]string); ok {
			style.Fill.Color = colors
		}
	}

	// Border
	if _, ok := options["border"]; ok {
		style.Border = []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		}
	}

	// Alignment
	if alignment, ok := options["alignment"].(map[string]interface{}); ok {
		style.Alignment = &excelize.Alignment{}
		if horizontal, ok := alignment["horizontal"].(string); ok {
			style.Alignment.Horizontal = horizontal
		}
		if vertical, ok := alignment["vertical"].(string); ok {
			style.Alignment.Vertical = vertical
		}
		if wrapText, ok := alignment["wrap_text"].(bool); ok {
			style.Alignment.WrapText = wrapText
		}
	}

	return f.NewStyle(style)
}

// ApplyHeaderStyle aplica un estilo predefinido para encabezados
func ApplyHeaderStyle(f *excelize.File, sheetName, cellRange string) error {
	style, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#E6E6FA"},
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	if err != nil {
		return err
	}

	return f.SetCellStyle(sheetName, cellRange, cellRange, style)
}

// ApplyErrorStyle aplica un estilo para celdas con errores
func ApplyErrorStyle(f *excelize.File, sheetName, cellRange string) error {
	style, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#FFE6E6"},
		},
		Border: []excelize.Border{
			{Type: "left", Color: "FF0000", Style: 1},
			{Type: "top", Color: "FF0000", Style: 1},
			{Type: "bottom", Color: "FF0000", Style: 1},
			{Type: "right", Color: "FF0000", Style: 1},
		},
	})

	if err != nil {
		return err
	}

	return f.SetCellStyle(sheetName, cellRange, cellRange, style)
}

// AutoFitColumns ajusta automáticamente el ancho de las columnas
func AutoFitColumns(f *excelize.File, sheetName string, columns []string, widths []float64) error {
	for i, col := range columns {
		if i < len(widths) {
			if err := f.SetColWidth(sheetName, col, col, widths[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

// AddDataValidation agrega validación de datos a un rango de celdas
func AddDataValidation(f *excelize.File, sheetName, cellRange string, validationType string, options []string) error {
	dv := excelize.NewDataValidation(true)

	switch validationType {
	case "list":
		dv.SetDropList(options)
	case "whole":
		if len(options) >= 2 {
			dv.SetRange(options[0], options[1], excelize.DataValidationTypeWhole, excelize.DataValidationOperatorBetween)
		}
	case "decimal":
		if len(options) >= 2 {
			dv.SetRange(options[0], options[1], excelize.DataValidationTypeDecimal, excelize.DataValidationOperatorBetween)
		}
	}

	dv.SetError(excelize.DataValidationErrorStyleStop, "Error de validación", "El valor ingresado no es válido")

	return f.AddDataValidation(sheetName, dv)
}

// GetUsedRange obtiene el rango de celdas utilizadas en una hoja
func GetUsedRange(f *excelize.File, sheetName string) (string, error) {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return "", err
	}

	if len(rows) == 0 {
		return "", fmt.Errorf("hoja vacía")
	}

	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	startCell := "A1"
	endCell := fmt.Sprintf("%s%d", IndexToExcelColumn(maxCols-1), len(rows))

	return fmt.Sprintf("%s:%s", startCell, endCell), nil
}
