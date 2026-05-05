package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/Diego-glitch-cloud/PY-WEB-BACKEND/db"
	"github.com/Diego-glitch-cloud/PY-WEB-BACKEND/handlers"
)

func main() {
	// Cargar variables de entorno desde .env (solo en desarrollo)
	if err := godotenv.Load(); err != nil {
		log.Println("no se encontro .env, usando variables de entorno del sistema")
	}

	// Conectar a PostgreSQL
	db.Connect()

	// Crear el router de Gin
	router := gin.Default()

	// Configurar CORS
	setupCORS(router)

	// Registrar rutas
	setupRoutes(router)

	// Iniciar servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("servidor corriendo en puerto %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("error al iniciar el servidor: %v", err)
	}
}

func setupCORS(router *gin.Engine) {
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type"},
	}))
}

func setupRoutes(router *gin.Engine) {
	// Archivos estaticos
	router.Static("/uploads", "./uploads")
	router.Static("/swagger", "./docs/swagger-ui")

	// Rutas de videojuegos
	router.GET("/games", handlers.GetGames)
	router.GET("/games/:id", handlers.GetGame)
	router.POST("/games", handlers.CreateGame)
	router.PUT("/games/:id", handlers.UpdateGame)
	router.DELETE("/games/:id", handlers.DeleteGame)

	// Rutas de ratings e imagen
	router.POST("/games/:id/rating", handlers.CreateRating)
	router.GET("/games/:id/rating", handlers.GetRatings)
	router.POST("/games/:id/image", handlers.UploadImage)
}
