package handlers

import (
	"client-data-compiler/internal/domain/models"
	"client-data-compiler/internal/services"
	"client-data-compiler/pkg/response"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	clientService services.ClientService
}

func NewUploadHandler(clientService services.ClientService) *UploadHandler {
	return &UploadHandler{
		clientService: clientService,
	}
}

// UploadExcel maneja la subida de archivos Excel
func (h *UploadHandler) UploadExcel(c *gin.Context) {
	log.Printf("Iniciando subida de archivo...")

	// Obtener archivo del formulario
	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("Error obteniendo archivo del formulario: %v", err)
		response.Error(c, http.StatusBadRequest, "No se proporcionó un archivo válido")
		return
	}

	log.Printf("Archivo recibido: %s, tamaño: %d bytes", file.Filename, file.Size)

	// Validar que el archivo no esté vacío
	if file.Size == 0 {
		log.Printf("Archivo vacío recibido")
		response.Error(c, http.StatusBadRequest, "El archivo está vacío")
		return
	}

	// Validar extensión del archivo
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".xlsx") {
		log.Printf("Extensión de archivo inválida: %s", file.Filename)
		response.Error(c, http.StatusBadRequest, "Solo se permiten archivos Excel (.xlsx)")
		return
	}

	// Validar tamaño del archivo (máximo 32MB)
	maxSize := int64(32 << 20) // 32 MB
	if file.Size > maxSize {
		log.Printf("Archivo demasiado grande: %d bytes", file.Size)
		response.Error(c, http.StatusBadRequest, "El archivo es demasiado grande. Tamaño máximo: 32MB")
		return
	}

	// Crear directorio uploads si no existe
	uploadsDir := "uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Printf("Error creando directorio uploads: %v", err)
		response.Error(c, http.StatusInternalServerError, "Error del servidor")
		return
	}

	// Generar nombre único para el archivo
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s", timestamp, file.Filename)

	// Limpiar nombre del archivo
	filename = sanitizeFilename(filename)

	// Ruta de destino
	uploadPath := filepath.Join(uploadsDir, filename)

	log.Printf("Guardando archivo en: %s", uploadPath)

	// Guardar archivo en el servidor
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		log.Printf("Error guardando archivo: %v", err)
		response.Error(c, http.StatusInternalServerError, "Error guardando archivo: "+err.Error())
		return
	}

	log.Printf("Archivo guardado exitosamente, procesando...")

	// Cargar y procesar el archivo Excel
	clients, err := h.clientService.LoadClientsFromExcel(uploadPath)
	if err != nil {
		log.Printf("Error procesando archivo Excel: %v", err)
		// Eliminar archivo si hay error en el procesamiento
		os.Remove(uploadPath)
		response.Error(c, http.StatusInternalServerError, fmt.Sprintf("Error procesando archivo: %v", err))
		return
	}

	log.Printf("Archivo procesado exitosamente: %d clientes cargados", len(clients))

	// Obtener estadísticas
	stats, _ := h.clientService.GetStats()

	// Preparar respuesta
	responseData := gin.H{
		"filename":        file.Filename,
		"uploaded_file":   filename,
		"total_clients":   len(clients),
		"valid_clients":   stats.Valid,
		"invalid_clients": stats.Invalid,
		"stats":           stats,
		"preview":         getPreviewClients(clients, 5), // Mostrar primeros 5 clientes
	}

	log.Printf("Respuesta preparada exitosamente")
	response.Success(c, "Archivo Excel cargado y procesado exitosamente", responseData)
}

// UploadMultiple maneja la subida de múltiples archivos Excel
func (h *UploadHandler) UploadMultiple(c *gin.Context) {
	log.Printf("Iniciando subida múltiple de archivos...")

	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("Error procesando formulario múltiple: %v", err)
		response.Error(c, http.StatusBadRequest, "Error procesando formulario: "+err.Error())
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		log.Printf("No se proporcionaron archivos")
		response.Error(c, http.StatusBadRequest, "No se proporcionaron archivos")
		return
	}

	log.Printf("Recibidos %d archivos para procesar", len(files))

	var results []gin.H
	var totalClients int
	var totalValid int
	var totalInvalid int

	// Crear directorio uploads si no existe
	uploadsDir := "uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Printf("Error creando directorio uploads: %v", err)
		response.Error(c, http.StatusInternalServerError, "Error del servidor")
		return
	}

	// Procesar cada archivo
	for i, file := range files {
		log.Printf("Procesando archivo %d/%d: %s", i+1, len(files), file.Filename)

		// Validar archivo
		if !strings.HasSuffix(strings.ToLower(file.Filename), ".xlsx") {
			log.Printf("Archivo %s tiene extensión inválida", file.Filename)
			results = append(results, gin.H{
				"filename": file.Filename,
				"status":   "error",
				"message":  "Solo se permiten archivos Excel (.xlsx)",
			})
			continue
		}

		if file.Size == 0 {
			log.Printf("Archivo %s está vacío", file.Filename)
			results = append(results, gin.H{
				"filename": file.Filename,
				"status":   "error",
				"message":  "El archivo está vacío",
			})
			continue
		}

		// Generar nombre único
		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("%s_%s", timestamp, file.Filename)
		filename = sanitizeFilename(filename)
		uploadPath := filepath.Join(uploadsDir, filename)

		// Guardar archivo
		if err := c.SaveUploadedFile(file, uploadPath); err != nil {
			log.Printf("Error guardando archivo %s: %v", file.Filename, err)
			results = append(results, gin.H{
				"filename": file.Filename,
				"status":   "error",
				"message":  "Error guardando archivo: " + err.Error(),
			})
			continue
		}

		// Procesar archivo
		clients, err := h.clientService.LoadClientsFromExcel(uploadPath)
		if err != nil {
			log.Printf("Error procesando archivo %s: %v", file.Filename, err)
			os.Remove(uploadPath)
			results = append(results, gin.H{
				"filename": file.Filename,
				"status":   "error",
				"message":  err.Error(),
			})
			continue
		}

		// Obtener estadísticas del archivo actual
		stats, _ := h.clientService.GetStats()

		results = append(results, gin.H{
			"filename":      file.Filename,
			"status":        "success",
			"total_clients": len(clients),
			"valid":         stats.Valid,
			"invalid":       stats.Invalid,
		})

		totalClients += len(clients)
		totalValid += stats.Valid
		totalInvalid += stats.Invalid

		log.Printf("Archivo %s procesado: %d clientes (%d válidos, %d inválidos)",
			file.Filename, len(clients), stats.Valid, stats.Invalid)
	}

	responseData := gin.H{
		"files_processed": len(files),
		"results":         results,
		"total_clients":   totalClients,
		"total_valid":     totalValid,
		"total_invalid":   totalInvalid,
	}

	log.Printf("Procesamiento múltiple completado: %d archivos procesados", len(files))
	response.Success(c, "Procesamiento de archivos completado", responseData)
}

// DownloadTemplate descarga una plantilla de Excel con la estructura correcta
func (h *UploadHandler) DownloadTemplate(c *gin.Context) {
	// Crear directorio templates si no existe
	templatesDir := "templates"
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		log.Printf("Error creando directorio templates: %v", err)
		response.Error(c, http.StatusInternalServerError, "Error del servidor")
		return
	}

	templatePath := filepath.Join(templatesDir, "plantilla_clientes.xlsx")

	// Si no existe la plantilla, crear una básica
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		if err := h.createTemplate(templatePath); err != nil {
			log.Printf("Error creando plantilla: %v", err)
			response.Error(c, http.StatusInternalServerError, "Error creando plantilla: "+err.Error())
			return
		}
	}

	// Verificar que el archivo existe
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		response.Error(c, http.StatusNotFound, "Plantilla no encontrada")
		return
	}

	// Establecer headers para descarga
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename=plantilla_clientes.xlsx")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	// Servir archivo
	c.File(templatePath)
}

// GetUploadedFiles obtiene la lista de archivos subidos
func (h *UploadHandler) GetUploadedFiles(c *gin.Context) {
	uploadsDir := "uploads"

	files, err := os.ReadDir(uploadsDir)
	if err != nil {
		log.Printf("Error leyendo directorio uploads: %v", err)
		response.Error(c, http.StatusInternalServerError, "Error leyendo directorio de archivos")
		return
	}

	var fileList []gin.H
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".xlsx") {
			info, _ := file.Info()
			fileList = append(fileList, gin.H{
				"name":          file.Name(),
				"size":          info.Size(),
				"modified_date": info.ModTime(),
				"download_url":  "/files/" + file.Name(),
			})
		}
	}

	response.Success(c, "Lista de archivos obtenida", gin.H{
		"files": fileList,
		"total": len(fileList),
	})
}

// DeleteUploadedFile elimina un archivo subido
func (h *UploadHandler) DeleteUploadedFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		response.Error(c, http.StatusBadRequest, "Nombre de archivo requerido")
		return
	}

	// Sanitizar nombre del archivo
	filename = sanitizeFilename(filename)
	filePath := filepath.Join("uploads", filename)

	// Verificar que el archivo existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		response.Error(c, http.StatusNotFound, "Archivo no encontrado")
		return
	}

	// Eliminar archivo
	if err := os.Remove(filePath); err != nil {
		log.Printf("Error eliminando archivo %s: %v", filePath, err)
		response.Error(c, http.StatusInternalServerError, "Error eliminando archivo: "+err.Error())
		return
	}

	response.Success(c, "Archivo eliminado exitosamente", gin.H{
		"filename": filename,
	})
}

// Funciones auxiliares

// sanitizeFilename limpia el nombre del archivo para evitar problemas de seguridad
func sanitizeFilename(filename string) string {
	// Reemplazar caracteres problemáticos
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, "..", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	filename = strings.ReplaceAll(filename, "*", "_")
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "\"", "_")
	filename = strings.ReplaceAll(filename, "<", "_")
	filename = strings.ReplaceAll(filename, ">", "_")
	filename = strings.ReplaceAll(filename, "|", "_")

	return filename
}

// getPreviewClients obtiene una vista previa de los primeros N clientes
func getPreviewClients(clients []*models.Client, limit int) []*models.Client {
	if len(clients) <= limit {
		return clients
	}
	return clients[:limit]
}

// createTemplate crea una plantilla básica de Excel
func (h *UploadHandler) createTemplate(templatePath string) error {
	// Crear directorio si no existe
	dir := filepath.Dir(templatePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Crear clientes de ejemplo para la plantilla
	sampleClients := []*models.Client{
		{
			ID:       1,
			Clave:    "1001",
			Nombre:   "Juan Pérez García",
			Correo:   "juan.perez@gmail.com",
			Telefono: "9611234567",
			IsValid:  true,
		},
		{
			ID:       2,
			Clave:    "1002",
			Nombre:   "María López Hernández",
			Correo:   "maria.lopez@hotmail.com",
			Telefono: "9629876543",
			IsValid:  true,
		},
		{
			ID:       3,
			Clave:    "1003",
			Nombre:   "Carlos Rodríguez Méndez",
			Correo:   "carlos.rodriguez@yahoo.com",
			Telefono: "9675551234",
			IsValid:  true,
		},
	}

	// Usar el servicio Excel para crear la plantilla
	excelService := services.NewExcelService()
	return excelService.WriteExcelFile(sampleClients, templatePath)
}
