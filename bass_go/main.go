package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"main.go/project/handlers"
	"main.go/project/models"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	err = db.AutoMigrate(&models.Project{}, &models.API{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("✅ Database connected and migrated successfully")

	projectHandler := handlers.NewProjectHandler(db)

	router := setupRouter(projectHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📋 Project APIs available at http://localhost:%s/admin/projects", port)

	log.Fatal(router.Run(":" + port))
}

func setupRouter(projectHandler *handlers.ProjectHandler) *gin.Engine {
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(corsMiddleware())

	baas := router.Group("baas")
	{
		baas.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "healthy",
				"service": "Backend Automation Service",
				"version": "1.0.0",
			})
		})
		baas.POST("/projects", projectHandler.CreateProject)
		baas.GET("/projects", projectHandler.ListProjects)
		baas.GET("/projects/:id", projectHandler.GetProject)
		baas.PUT("/projects/:id", projectHandler.UpdateProject)
		baas.DELETE("/projects/:id", projectHandler.DeleteProject)
	}
	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
