package services

import (
	"client-data-compiler/internal/domain/models"
	"client-data-compiler/internal/utils"
	"sync"
)

type ValidationService interface {
	ValidateClient(client *models.Client) *models.Client
	ValidateClients(clients []*models.Client) []*models.Client
	ValidateClientsConcurrent(clients []*models.Client) []*models.Client
}

type validationService struct{}

func NewValidationService() ValidationService {
	return &validationService{}
}

// ValidateClient valida un cliente individual
func (s *validationService) ValidateClient(client *models.Client) *models.Client {
	// Limpiar errores previos
	client.ClearErrors()

	// Validar clave
	if valid, msg := utils.ValidateClientKey(client.Clave); !valid {
		client.AddError("clave", msg)
	}

	// Validar nombre
	client.Nombre = utils.CleanString(client.Nombre)
	if valid, msg := utils.ValidateClientName(client.Nombre); !valid {
		client.AddError("nombre", msg)
	}

	// Validar correo
	client.Correo = utils.CleanString(client.Correo)
	if valid, msg := utils.ValidateEmail(client.Correo); !valid {
		client.AddError("correo", msg)
	}

	// Validar teléfono
	client.Telefono = utils.CleanString(client.Telefono)
	if valid, msg := utils.ValidatePhone(client.Telefono); !valid {
		client.AddError("telefono", msg)
	}

	// Actualizar estado de validez
	client.IsValid = len(client.Errors) == 0

	return client
}

// ValidateClients valida múltiples clientes secuencialmente
func (s *validationService) ValidateClients(clients []*models.Client) []*models.Client {
	validatedClients := make([]*models.Client, len(clients))

	for i, client := range clients {
		validatedClients[i] = s.ValidateClient(client)
	}

	return validatedClients
}

// ValidateClientsConcurrent valida múltiples clientes usando goroutines para mejor rendimiento
func (s *validationService) ValidateClientsConcurrent(clients []*models.Client) []*models.Client {
	if len(clients) == 0 {
		return clients
	}

	// Para pocos clientes, usar validación secuencial
	if len(clients) < 100 {
		return s.ValidateClients(clients)
	}

	// Usar workers para validación concurrente
	numWorkers := 10
	if len(clients) < numWorkers {
		numWorkers = len(clients)
	}

	jobs := make(chan int, len(clients))
	var wg sync.WaitGroup

	// Crear workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				clients[index] = s.ValidateClient(clients[index])
			}
		}()
	}

	// Enviar trabajos
	for i := range clients {
		jobs <- i
	}
	close(jobs)

	// Esperar a que terminen todos los workers
	wg.Wait()

	return clients
}

// GetValidationStats obtiene estadísticas de validación
func (s *validationService) GetValidationStats(clients []*models.Client) *models.ClientStats {
	stats := &models.ClientStats{
		Total:         len(clients),
		Valid:         0,
		Invalid:       0,
		ErrorsByField: make(map[string]int),
	}

	for _, client := range clients {
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

	return stats
}
