package main

import (
	"context"
	"log"
	"movie-app-go/database"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
		router := gin.Default()

		router.GET("/", func(c *gin.Context) {
			c.String(200, "Hello to  My movie app Server!")
		})

		err:=godotenv.Load(".env")
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
			err:= client.Disconnect(context.Background())
			if err != nil {
				log.Fatalf("Error disconnecting from MongoDB: %v", err)
			}
		}()

		if err:= router.Run(":5000"); err != nil {
			log.Fatalf("Failed to run server: %v", err)
		}
	}