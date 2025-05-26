package services

import (
	"client-data-compiler/internal/domain/errors"
	"client-data-compiler/internal/domain/models"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ClientService interface {
	LoadClientsFromExcel(filePath string) ([]*models.Client, error)
	GetClients(filter *models.ClientFilter) ([]*models.Client, error)
	GetClientByID(id int) (*models.Client, error)
	UpdateClient(id int, client *models.Client) (*models.Client, error)
	DeleteClient(id int) error
	ValidateAllClients() ([]*models.Client, error)
	ValidateClient(client *models.Client) *models.Client
	ExportClientsToExcel(filename string) (string, error)
	GetStats() (*models.ClientStats, error)
	ClearAllClients() error
	GetClientCount() int
}

type clientService struct {
	clients           []*models.Client
	mu                sync.RWMutex
	excelService      ExcelService
	validationService ValidationService
	lastID            int
}

func NewClientService(excelService ExcelService, validationService ValidationService) ClientService {
	return &clientService{
		clients:           make([]*models.Client, 0),
		excelService:      excelService,
		validationService: validationService,
		lastID:            0,
	}
}

// LoadClientsFromExcel carga clientes desde un archivo Excel
func (s *clientService) LoadClientsFromExcel(filePath string) ([]*models.Client, error) {
	// Validar estructura del archivo
	if err := s.excelService.ValidateExcelStructure(filePath); err != nil {
		return nil, err
	}

	// Leer archivo Excel
	clients, err := s.excelService.ReadExcelFile(filePath)
	if err != nil {
		return nil, err
	}

	// Validar clientes
	clients = s.validationService.ValidateClientsConcurrent(clients)

	// Verificar claves duplicadas
	s.checkDuplicateKeys(clients)

	// Almacenar en memoria
	s.mu.Lock()
	s.clients = clients
	s.updateLastID()
	s.mu.Unlock()

	return clients, nil
}

// GetClients obtiene clientes con filtros opcionales
func (s *clientService) GetClients(filter *models.ClientFilter) ([]*models.Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if filter == nil {
		return s.clients, nil
	}

	// Aplicar filtros
	filteredClients := make([]*models.Client, 0)

	for _, client := range s.clients {
		if s.matchesFilter(client, filter) {
			filteredClients = append(filteredClients, client)
		}
	}

	// Aplicar paginación
	if filter.Page > 0 && filter.Limit > 0 {
		start := (filter.Page - 1) * filter.Limit
		end := start + filter.Limit

		if start >= len(filteredClients) {
			return []*models.Client{}, nil
		}

		if end > len(filteredClients) {
			end = len(filteredClients)
		}

		return filteredClients[start:end], nil
	}

	return filteredClients, nil
}

// GetClientByID obtiene un cliente por su ID
func (s *clientService) GetClientByID(id int) (*models.Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		if client.ID == id {
			return client, nil
		}
	}

	return nil, errors.ErrClientNotFound
}

// UpdateClient actualiza un cliente existente
func (s *clientService) UpdateClient(id int, updatedClient *models.Client) (*models.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Buscar cliente
	clientIndex := -1
	for i, client := range s.clients {
		if client.ID == id {
			clientIndex = i
			break
		}
	}

	if clientIndex == -1 {
		return nil, errors.ErrClientNotFound
	}

	// Verificar clave duplicada (excluyendo el cliente actual)
	for i, client := range s.clients {
		if i != clientIndex && client.Clave == updatedClient.Clave {
			return nil, errors.ErrDuplicateClientKey
		}
	}

	// Mantener datos originales
	originalClient := s.clients[clientIndex]
	updatedClient.ID = originalClient.ID
	updatedClient.RowNumber = originalClient.RowNumber
	updatedClient.CreatedAt = originalClient.CreatedAt
	updatedClient.UpdatedAt = time.Now()

	// Validar cliente actualizado
	validatedClient := s.validationService.ValidateClient(updatedClient)

	// Actualizar en memoria
	s.clients[clientIndex] = validatedClient

	return validatedClient, nil
}

// DeleteClient elimina un cliente
func (s *clientService) DeleteClient(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	clientIndex := -1
	for i, client := range s.clients {
		if client.ID == id {
			clientIndex = i
			break
		}
	}

	if clientIndex == -1 {
		return errors.ErrClientNotFound
	}

	// Eliminar cliente
	s.clients = append(s.clients[:clientIndex], s.clients[clientIndex+1:]...)

	return nil
}

// ValidateAllClients valida todos los clientes cargados
func (s *clientService) ValidateAllClients() ([]*models.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validar todos los clientes
	s.clients = s.validationService.ValidateClientsConcurrent(s.clients)

	// Verificar claves duplicadas
	s.checkDuplicateKeys(s.clients)

	return s.clients, nil
}

// ValidateClient valida un cliente individual
func (s *clientService) ValidateClient(client *models.Client) *models.Client {
	return s.validationService.ValidateClient(client)
}

// ExportClientsToExcel exporta los clientes a un archivo Excel
func (s *clientService) ExportClientsToExcel(filename string) (string, error) {
	s.mu.RLock()
	clients := make([]*models.Client, len(s.clients))
	copy(clients, s.clients)
	s.mu.RUnlock()

	if len(clients) == 0 {
		return "", errors.NewFileProcessingError("No hay clientes para exportar")
	}

	// Generar nombre de archivo único si no se proporciona
	if filename == "" {
		timestamp := time.Now().Format("20060102_150405")
		filename = fmt.Sprintf("clientes_exportados_%s.xlsx", timestamp)
	}

	// Asegurar que termine en .xlsx
	if !strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
		filename += ".xlsx"
	}

	// Ruta completa del archivo
	filePath := filepath.Join("uploads", filename)

	// Exportar a Excel
	if err := s.excelService.WriteExcelFile(clients, filePath); err != nil {
		return "", err
	}

	return filePath, nil
}

// GetStats obtiene estadísticas de los clientes
func (s *clientService) GetStats() (*models.ClientStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &models.ClientStats{
		Total:         len(s.clients),
		Valid:         0,
		Invalid:       0,
		ErrorsByField: make(map[string]int),
	}

	for _, client := range s.clients {
		if client.IsValid {
			stats.Valid++
		} else {
			stats.Invalid++

			// Contar errores por campo
			for field := range client.Errors {
				stats.ErrorsByField[field]++
			}
		}
	}

	return stats, nil
}

// ClearAllClients limpia todos los clientes de la memoria
func (s *clientService) ClearAllClients() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients = make([]*models.Client, 0)
	s.lastID = 0

	return nil
}

// GetClientCount obtiene el número total de clientes
func (s *clientService) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.clients)
}

// Métodos auxiliares privados

// matchesFilter verifica si un cliente coincide con los filtros
func (s *clientService) matchesFilter(client *models.Client, filter *models.ClientFilter) bool {
	// Filtro por clave
	if filter.Clave != "" {
		if !strings.Contains(strings.ToLower(client.Clave), strings.ToLower(filter.Clave)) {
			return false
		}
	}

	// Filtro por nombre
	if filter.Nombre != "" {
		if !strings.Contains(strings.ToLower(client.Nombre), strings.ToLower(filter.Nombre)) {
			return false
		}
	}

	// Filtro por correo
	if filter.Correo != "" {
		if !strings.Contains(strings.ToLower(client.Correo), strings.ToLower(filter.Correo)) {
			return false
		}
	}

	// Filtro por teléfono
	if filter.Telefono != "" {
		if !strings.Contains(client.Telefono, filter.Telefono) {
			return false
		}
	}

	// Filtro por estado de validación
	if filter.HasErrors != nil {
		hasErrors := !client.IsValid
		if *filter.HasErrors != hasErrors {
			return false
		}
	}

	return true
}

// checkDuplicateKeys verifica y marca claves duplicadas
func (s *clientService) checkDuplicateKeys(clients []*models.Client) {
	keyCount := make(map[string][]int)

	// Contar ocurrencias de cada clave
	for i, client := range clients {
		if client.Clave != "" {
			keyCount[client.Clave] = append(keyCount[client.Clave], i)
		}
	}

	// Marcar duplicados
	for key, indices := range keyCount {
		if len(indices) > 1 {
			for _, index := range indices {
				clients[index].AddError("clave", fmt.Sprintf("Clave duplicada: %s", key))
			}
		}
	}
}

// updateLastID actualiza el último ID usado
func (s *clientService) updateLastID() {
	maxID := 0
	for _, client := range s.clients {
		if client.ID > maxID {
			maxID = client.ID
		}
	}
	s.lastID = maxID
}

// CleanupTempFiles limpia archivos temporales antiguos
func (s *clientService) CleanupTempFiles() error {
	uploadsDir := "uploads"

	return filepath.Walk(uploadsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continuar con otros archivos
		}

		// Eliminar archivos más antiguos de 24 horas
		if time.Since(info.ModTime()) > 24*time.Hour {
			os.Remove(path)
		}

		return nil
	})
}
