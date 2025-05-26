// cmd/api/main.go
package main

import (
	"client-data-compiler/internal/config"
	"client-data-compiler/internal/handlers"
	"client-data-compiler/internal/services"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Cargar configuración
	cfg := config.Load()

	// Crear directorios necesarios
	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Fatal("Error creando directorio uploads:", err)
	}
	if err := os.MkdirAll("templates", 0755); err != nil {
		log.Fatal("Error creando directorio templates:", err)
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

	// Configurar CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:5173",
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Content-Length",
			"Accept-Encoding", "X-CSRF-Token", "Authorization",
			"accept", "origin", "Cache-Control", "X-Requested-With",
		},
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	// Middleware adicional
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Configurar límite de tamaño de archivo
	router.MaxMultipartMemory = 32 << 20 // 32 MiB

	// Health check
	router.GET("/health", func(c *gin.Context) {
		log.Printf("✅ Health check desde: %s", c.Request.Header.Get("Origin"))
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Server is running",
			"port":    cfg.Port,
		})
	})

	// Test endpoint
	router.GET("/test", func(c *gin.Context) {
		log.Printf("🧪 Test endpoint accessed")
		c.JSON(http.StatusOK, gin.H{
			"message": "Backend is working correctly",
			"cors":    "enabled",
		})
	})

	// Rutas API - SIN trailing slash en los grupos
	api := router.Group("/api")
	{
		// Upload de archivos
		api.POST("/upload", uploadHandler.UploadExcel)
		api.POST("/upload/multiple", uploadHandler.UploadMultiple)
		api.GET("/upload/template", uploadHandler.DownloadTemplate)
		api.GET("/upload/files", uploadHandler.GetUploadedFiles)
		api.DELETE("/upload/files/:filename", uploadHandler.DeleteUploadedFile)

		// Gestión de clientes
		api.GET("/clients", clientHandler.GetClients)
		api.GET("/clients/search", clientHandler.SearchClients)
		api.GET("/clients/:id", clientHandler.GetClientByID)
		api.PUT("/clients/:id", clientHandler.UpdateClient)
		api.DELETE("/clients/:id", clientHandler.DeleteClient)
		api.DELETE("/clients", clientHandler.ClearAll)

		// Validaciones
		api.GET("/validate", clientHandler.ValidateAll)
		api.POST("/validate/single", clientHandler.ValidateSingle)

		// Exportar y estadísticas
		api.GET("/export", clientHandler.ExportExcel)
		api.GET("/stats", clientHandler.GetStats)
	}

	// Servir archivos estáticos
	router.Static("/files", "./uploads")

	log.Printf("🚀 Servidor iniciado en puerto %s", cfg.Port)
	log.Printf("🌐 CORS configurado para: http://localhost:3000")
	log.Printf("📍 Health check: http://localhost:%s/health", cfg.Port)
	log.Printf("📋 API base: http://localhost:%s/api", cfg.Port)
	log.Printf("🧪 Test: http://localhost:%s/test", cfg.Port)
	log.Printf("👥 Clientes: http://localhost:%s/api/clients", cfg.Port)

	log.Fatal(router.Run(":" + cfg.Port))
}
