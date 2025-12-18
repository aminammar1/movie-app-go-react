package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"movie-app-go/database"
	_ "movie-app-go/docs"
	"movie-app-go/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// @title Movie API
// @version 1.0
// @description API for the movie app.
// @BasePath /api/v1
// @schemes http
// @host localhost:5000
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.String(200, "Hello to  My movie api!")
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")

	var origins []string
	if allowedOrigins != "" {
		origins = strings.Split(allowedOrigins, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
			log.Println("Allowed Origin:", origins[i])
		}
	} else {
		origins = []string{"http://localhost:5173"}
		log.Println("No ALLOWED_ORIGINS set, defaulting to http://localhost:5173")
	}

	corsConfig := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	router.Use(cors.New(corsConfig))
	router.Use(gin.Logger())

	var client *mongo.Client = database.ConnectDB()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	log.Println("Connected to MongoDB successfully")

	defer func() {
		err := client.Disconnect(context.Background())
		if err != nil {
			log.Fatalf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	routes.SetupPublicRoutes(router, client)
	routes.SetupProtectedRoutes(router, client)

	if err := router.Run(":5000"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
