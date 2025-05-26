package handlers

import (
	"client-data-compiler/internal/domain/models"
	"client-data-compiler/internal/services"
	"client-data-compiler/pkg/response"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ClientHandler struct {
	clientService services.ClientService
}

func NewClientHandler(clientService services.ClientService) *ClientHandler {
	return &ClientHandler{
		clientService: clientService,
	}
}

// GetClients obtiene clientes con filtros opcionales
func (h *ClientHandler) GetClients(c *gin.Context) {
	log.Printf("📥 Petición GetClients recibida desde: %s", c.Request.Header.Get("Origin"))
	log.Printf("📊 Query params: %s", c.Request.URL.RawQuery)

	// Construir filtros desde query parameters
	filter := &models.ClientFilter{
		Clave:    c.Query("clave"),
		Nombre:   c.Query("nombre"),
		Correo:   c.Query("correo"),
		Telefono: c.Query("telefono"),
	}

	// Filtro por errores
	if hasErrorsStr := c.Query("has_errors"); hasErrorsStr != "" {
		if hasErrors, err := strconv.ParseBool(hasErrorsStr); err == nil {
			filter.HasErrors = &hasErrors
		}
	}

	// Paginación
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	log.Printf("🔍 Filtros aplicados: %+v", filter)

	// Obtener clientes
	clients, err := h.clientService.GetClients(filter)
	if err != nil {
		log.Printf("❌ Error obteniendo clientes: %v", err)
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Obtener total de clientes (sin filtros para paginación)
	totalClients := h.clientService.GetClientCount()

	log.Printf("✅ Enviando respuesta: %d clientes encontrados", len(clients))

	responseData := gin.H{
		"clients": clients,
		"total":   totalClients,
		"page":    filter.Page,
		"limit":   filter.Limit,
	}

	response.Success(c, "Clientes obtenidos exitosamente", responseData)
}

// SearchClients busca clientes por texto libre
func (h *ClientHandler) SearchClients(c *gin.Context) {
	searchTerm := c.Query("q")
	if searchTerm == "" {
		response.Error(c, http.StatusBadRequest, "Término de búsqueda requerido")
		return
	}

	// Para búsqueda libre, obtenemos todos los clientes y filtramos manualmente
	allClients, err := h.clientService.GetClients(nil)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Buscar en todos los campos
	var results []*models.Client
	searchTermLower := strings.ToLower(searchTerm)

	for _, client := range allClients {
		if strings.Contains(strings.ToLower(client.Clave), searchTermLower) ||
			strings.Contains(strings.ToLower(client.Nombre), searchTermLower) ||
			strings.Contains(strings.ToLower(client.Correo), searchTermLower) ||
			strings.Contains(client.Telefono, searchTerm) {
			results = append(results, client)
		}
	}

	response.Success(c, "Búsqueda completada", gin.H{
		"clients":     results,
		"total":       len(results),
		"search_term": searchTerm,
	})
}

// GetClientByID obtiene un cliente específico por ID
func (h *ClientHandler) GetClientByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID de cliente inválido")
		return
	}

	client, err := h.clientService.GetClientByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, "Cliente encontrado", gin.H{"client": client})
}

// UpdateClient actualiza un cliente existente
func (h *ClientHandler) UpdateClient(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID de cliente inválido")
		return
	}

	var updateData models.Client
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos de cliente inválidos: "+err.Error())
		return
	}

	updatedClient, err := h.clientService.UpdateClient(id, &updateData)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, "Cliente actualizado exitosamente", gin.H{"client": updatedClient})
}

// DeleteClient elimina un cliente
func (h *ClientHandler) DeleteClient(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID de cliente inválido")
		return
	}

	if err := h.clientService.DeleteClient(id); err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, "Cliente eliminado exitosamente", nil)
}

// ClearAll limpia todos los clientes de la memoria
func (h *ClientHandler) ClearAll(c *gin.Context) {
	if err := h.clientService.ClearAllClients(); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, "Todos los clientes han sido eliminados", nil)
}

// ValidateAll valida todos los clientes cargados
func (h *ClientHandler) ValidateAll(c *gin.Context) {
	clients, err := h.clientService.ValidateAllClients()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Obtener estadísticas de validación
	stats, _ := h.clientService.GetStats()

	responseData := gin.H{
		"clients": clients,
		"stats":   stats,
	}

	response.Success(c, "Validación completada", responseData)
}

// ValidateSingle valida un cliente individual
func (h *ClientHandler) ValidateSingle(c *gin.Context) {
	var clientData models.Client
	if err := c.ShouldBindJSON(&clientData); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos de cliente inválidos: "+err.Error())
		return
	}

	validatedClient := h.clientService.ValidateClient(&clientData)

	response.Success(c, "Cliente validado", gin.H{"client": validatedClient})
}

// ExportExcel exporta los clientes a un archivo Excel
func (h *ClientHandler) ExportExcel(c *gin.Context) {
	filename := c.Query("filename")

	filePath, err := h.clientService.ExportClientsToExcel(filename)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	responseData := gin.H{
		"file_path": filePath,
		"file_url":  "/files/" + filePath[8:], // Remover "uploads/" del path
	}

	response.Success(c, "Archivo Excel exportado exitosamente", responseData)
}

// GetStats obtiene estadísticas de los clientes
func (h *ClientHandler) GetStats(c *gin.Context) {
	stats, err := h.clientService.GetStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, "Estadísticas obtenidas exitosamente", gin.H{"stats": stats})
}
