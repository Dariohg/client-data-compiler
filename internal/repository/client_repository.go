package repository

import (
	"client-data-compiler/internal/domain/errors"
	"client-data-compiler/internal/domain/models"
	"strings"
	"sync"
	"time"
)

// ClientRepository interfaz para el repositorio de clientes
type ClientRepository interface {
	Create(client *models.Client) (*models.Client, error)
	GetByID(id int) (*models.Client, error)
	GetByClave(clave string) (*models.Client, error)
	GetAll() ([]*models.Client, error)
	Update(id int, client *models.Client) (*models.Client, error)
	Delete(id int) error
	Clear() error
	Count() int
	FindByFilter(filter *models.ClientFilter) ([]*models.Client, error)
	BatchCreate(clients []*models.Client) ([]*models.Client, error)
	BatchUpdate(clients []*models.Client) ([]*models.Client, error)
	GetDuplicateKeys() map[string][]int
}

// inMemoryClientRepository implementación en memoria del repositorio
type inMemoryClientRepository struct {
	clients map[int]*models.Client
	mutex   sync.RWMutex
	lastID  int
}

// NewInMemoryClientRepository crea una nueva instancia del repositorio en memoria
func NewInMemoryClientRepository() ClientRepository {
	return &inMemoryClientRepository{
		clients: make(map[int]*models.Client),
		lastID:  0,
	}
}

// Create crea un nuevo cliente
func (r *inMemoryClientRepository) Create(client *models.Client) (*models.Client, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Verificar clave duplicada
	for _, existingClient := range r.clients {
		if existingClient.Clave == client.Clave {
			return nil, errors.ErrDuplicateClientKey
		}
	}

	// Asignar nuevo ID
	r.lastID++
	client.ID = r.lastID
	client.CreatedAt = time.Now()
	client.UpdatedAt = time.Now()

	// Guardar cliente
	r.clients[client.ID] = client

	return client, nil
}

// GetByID obtiene un cliente por su ID
func (r *inMemoryClientRepository) GetByID(id int) (*models.Client, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	client, exists := r.clients[id]
	if !exists {
		return nil, errors.ErrClientNotFound
	}

	return client, nil
}

// GetByClave obtiene un cliente por su clave
func (r *inMemoryClientRepository) GetByClave(clave string) (*models.Client, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, client := range r.clients {
		if client.Clave == clave {
			return client, nil
		}
	}

	return nil, errors.ErrClientNotFound
}

// GetAll obtiene todos los clientes
func (r *inMemoryClientRepository) GetAll() ([]*models.Client, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	clients := make([]*models.Client, 0, len(r.clients))
	for _, client := range r.clients {
		clients = append(clients, client)
	}

	return clients, nil
}

// Update actualiza un cliente existente
func (r *inMemoryClientRepository) Update(id int, updatedClient *models.Client) (*models.Client, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Verificar que el cliente existe
	existingClient, exists := r.clients[id]
	if !exists {
		return nil, errors.ErrClientNotFound
	}

	// Verificar clave duplicada (excluyendo el cliente actual)
	for clientID, client := range r.clients {
		if clientID != id && client.Clave == updatedClient.Clave {
			return nil, errors.ErrDuplicateClientKey
		}
	}

	// Mantener datos originales
	updatedClient.ID = existingClient.ID
	updatedClient.CreatedAt = existingClient.CreatedAt
	updatedClient.UpdatedAt = time.Now()

	// Actualizar cliente
	r.clients[id] = updatedClient

	return updatedClient, nil
}

// Delete elimina un cliente
func (r *inMemoryClientRepository) Delete(id int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.clients[id]; !exists {
		return errors.ErrClientNotFound
	}

	delete(r.clients, id)
	return nil
}

// Clear elimina todos los clientes
func (r *inMemoryClientRepository) Clear() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.clients = make(map[int]*models.Client)
	r.lastID = 0

	return nil
}

// Count obtiene el número total de clientes
func (r *inMemoryClientRepository) Count() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.clients)
}

// FindByFilter busca clientes por filtros
func (r *inMemoryClientRepository) FindByFilter(filter *models.ClientFilter) ([]*models.Client, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var results []*models.Client

	for _, client := range r.clients {
		if r.matchesFilter(client, filter) {
			results = append(results, client)
		}
	}

	// Aplicar paginación si está especificada
	if filter.Page > 0 && filter.Limit > 0 {
		start := (filter.Page - 1) * filter.Limit
		end := start + filter.Limit

		if start >= len(results) {
			return []*models.Client{}, nil
		}

		if end > len(results) {
			end = len(results)
		}

		results = results[start:end]
	}

	return results, nil
}

// BatchCreate crea múltiples clientes
func (r *inMemoryClientRepository) BatchCreate(clients []*models.Client) ([]*models.Client, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	createdClients := make([]*models.Client, 0, len(clients))

	for _, client := range clients {
		// Asignar nuevo ID
		r.lastID++
		client.ID = r.lastID
		client.CreatedAt = time.Now()
		client.UpdatedAt = time.Now()

		// Guardar cliente
		r.clients[client.ID] = client
		createdClients = append(createdClients, client)
	}

	return createdClients, nil
}

// BatchUpdate actualiza múltiples clientes
func (r *inMemoryClientRepository) BatchUpdate(clients []*models.Client) ([]*models.Client, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	updatedClients := make([]*models.Client, 0, len(clients))

	for _, client := range clients {
		if _, exists := r.clients[client.ID]; exists {
			client.UpdatedAt = time.Now()
			r.clients[client.ID] = client
			updatedClients = append(updatedClients, client)
		}
	}

	return updatedClients, nil
}

// GetDuplicateKeys obtiene las claves duplicadas
func (r *inMemoryClientRepository) GetDuplicateKeys() map[string][]int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	keyCount := make(map[string][]int)

	// Contar ocurrencias de cada clave
	for _, client := range r.clients {
		if client.Clave != "" {
			keyCount[client.Clave] = append(keyCount[client.Clave], client.ID)
		}
	}

	// Filtrar solo las duplicadas
	duplicates := make(map[string][]int)
	for key, ids := range keyCount {
		if len(ids) > 1 {
			duplicates[key] = ids
		}
	}

	return duplicates
}

// Métodos auxiliares privados

// matchesFilter verifica si un cliente coincide con los filtros
func (r *inMemoryClientRepository) matchesFilter(client *models.Client, filter *models.ClientFilter) bool {
	// Filtro por clave
	if filter.Clave != "" {
		if !r.containsIgnoreCase(client.Clave, filter.Clave) {
			return false
		}
	}

	// Filtro por nombre
	if filter.Nombre != "" {
		if !r.containsIgnoreCase(client.Nombre, filter.Nombre) {
			return false
		}
	}

	// Filtro por correo
	if filter.Correo != "" {
		if !r.containsIgnoreCase(client.Correo, filter.Correo) {
			return false
		}
	}

	// Filtro por teléfono
	if filter.Telefono != "" {
		if !r.containsIgnoreCase(client.Telefono, filter.Telefono) {
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

// containsIgnoreCase verifica si una cadena contiene otra (ignorando mayúsculas)
func (r *inMemoryClientRepository) containsIgnoreCase(haystack, needle string) bool {
	return len(haystack) >= len(needle) &&
		strings.ToLower(haystack) != strings.ToLower(haystack[len(needle):]) ||
		strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

// GetStats obtiene estadísticas del repositorio
func (r *inMemoryClientRepository) GetStats() (*models.ClientStats, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	stats := &models.ClientStats{
		Total:         len(r.clients),
		Valid:         0,
		Invalid:       0,
		ErrorsByField: make(map[string]int),
	}

	for _, client := range r.clients {
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
