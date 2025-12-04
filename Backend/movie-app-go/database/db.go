package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func ConnectDB() *mongo.Client {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mongoURI := os.Getenv("MONGO_URI")

	if mongoURI == "" {
		log.Fatal("MONGO_URI not set in environment")
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(nil, clientOptions)

	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}

	return client
}


func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {

	err:= godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		log.Fatal("DB_NAME not set in environment")
	}

	fmt.Println("Database Name :", dbName)

	collection := client.Database(dbName).Collection(collectionName)

	if collection == nil {
		return nil
	}

	return collection
}
