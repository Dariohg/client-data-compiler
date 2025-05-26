// cmd/api/main.go
package main

import (
	"client-data-compiler/internal/config"
	"client-data-compiler/internal/handlers"
	"client-data-compiler/internal/services"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// corsMiddleware es un middleware personalizado para CORS
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Permitir estos orígenes
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:5173",
		}

		// Verificar si el origen está permitido
		originAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				originAllowed = true
				break
			}
		}

		// Establecer headers CORS
		if originAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "http://localhost:3000") // fallback
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// Manejar preflight requests
		if c.Request.Method == "OPTIONS" {
			log.Printf("🔄 Preflight request para: %s %s", c.Request.Method, c.Request.URL.Path)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		log.Printf("📥 Request: %s %s desde %s", c.Request.Method, c.Request.URL.Path, origin)
		c.Next()
	}
}

func main() {
	// Cargar configuración
	cfg := config.Load()

	// Crear directorio de uploads si no existe
	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Fatal("Error creando directorio uploads:", err)
	}

	// Inicializar servicios
	excelService := services.NewExcelService()
	validationService := services.NewValidationService()
	clientService := services.NewClientService(excelService, validationService)

	// Inicializar handlers
	clientHandler := handlers.NewClientHandler(clientService)
	uploadHandler := handlers.NewUploadHandler(clientService)

	// Configurar Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Aplicar middleware CORS PRIMERO
	router.Use(corsMiddleware())

	// Configurar límite de tamaño de archivo
	router.MaxMultipartMemory = 32 << 20 // 32 MiB

	// Health check
	router.GET("/health", func(c *gin.Context) {
		log.Printf("✅ Health check desde: %s", c.Request.Header.Get("Origin"))
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Rutas API
	api := router.Group("/api")
	{
		// Upload de archivos
		upload := api.Group("/upload")
		{
			upload.POST("/", uploadHandler.UploadExcel)
			upload.POST("/multiple", uploadHandler.UploadMultiple)
			upload.GET("/template", uploadHandler.DownloadTemplate)
			upload.GET("/files", uploadHandler.GetUploadedFiles)
			upload.DELETE("/files/:filename", uploadHandler.DeleteUploadedFile)
		}

		// Gestión de clientes
		clients := api.Group("/clients")
		{
			clients.GET("/", clientHandler.GetClients)
			clients.GET("/search", clientHandler.SearchClients)
			clients.GET("/:id", clientHandler.GetClientByID)
			clients.PUT("/:id", clientHandler.UpdateClient)
			clients.DELETE("/:id", clientHandler.DeleteClient)
			clients.DELETE("/", clientHandler.ClearAll)
		}

		// Validaciones
		validate := api.Group("/validate")
		{
			validate.GET("/", clientHandler.ValidateAll)
			validate.POST("/single", clientHandler.ValidateSingle)
		}

		// Exportar
		api.GET("/export", clientHandler.ExportExcel)

		// Estadísticas
		api.GET("/stats", clientHandler.GetStats)
	}

	// Servir archivos estáticos (Excel exportados)
	router.Static("/files", "./uploads")

	log.Printf("🚀 Servidor iniciado en puerto %s", cfg.Port)
	log.Printf("🌐 CORS configurado para: http://localhost:3000")
	log.Printf("📍 Health check disponible en: http://localhost:%s/health", cfg.Port)
	log.Printf("📋 API disponible en: http://localhost:%s/api", cfg.Port)

	log.Fatal(router.Run(":" + cfg.Port))
}
