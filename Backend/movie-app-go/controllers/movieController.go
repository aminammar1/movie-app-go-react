package controllers

import (
	"net/http"
	"context"
	"errors"
	//"log"
	//"strconv"
	//"strings"
	"time"

	"movie-app-go/database"
	"movie-app-go/models"
	//"movie-app-go/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	//"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	//"go.mongodb.org/mongo-driver/v2/mongo/options"
)


var validate = validator.New()

func AddMovie(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movie models.Movie

		if err := c.ShouldBindJSON(&movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := validate.Struct(movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var movieCollection = database.OpenCollection(client, "movies")
		data, err := movieCollection.InsertOne(ctx, movie)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting movie"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": data})
	}
}

func GetMovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movieCollection = database.OpenCollection(client, "movies")

		cursor , err := movieCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching movies"})
			return
		}
		defer cursor.Close(ctx)

		var movies []models.Movie
		if err = cursor.All(ctx, &movies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while decoding movies"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": movies})
	}
}

func GetMovieByID(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		movieID := c.Param("imdb_id")

		if movieID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Movie ID is required"})
			return
		}

		var movieCollection = database.OpenCollection(client, "movies")

		var movie models.Movie
		err := movieCollection.FindOne(ctx, bson.M{"imdb_id": movieID}).Decode(&movie)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching movie"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": movie})
	}
}

func SearchMovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		queryParams := c.Request.URL.Query()
		filter := bson.M{}

		for key, values := range queryParams {
			if len(values) > 0 {
				filter[key] = bson.M{"$regex": values[0], "$options": "i"}
			}
		}

		var movieCollection = database.OpenCollection(client, "movies")

		cursor, err := movieCollection.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while searching movies"})
			return
		}
		defer cursor.Close(ctx)

		var movies []models.Movie
		if err = cursor.All(ctx, &movies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while decoding movies"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": movies})
	}
}

func GetRankings(client *mongo.Client , c *gin.Context)  ([]models.Ranking, error) {
	var rankins []models.Ranking
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var rankingCollection *mongo.Collection = database.OpenCollection(client, "rankings")
	cursor, err := rankingCollection.Find(ctx, bson.M{})
	if err !=nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &rankins); err != nil {
		return nil, err
	}

	return rankins, nil
}