package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"movie-app-go/database"
	"movie-app-go/models"
	"movie-app-go/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms/openai"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var validate = validator.New()

// @Summary Add a movie
// @Tags movies
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body models.Movie true "Movie"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /addmovie [post]
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

// @Summary List movies
// @Tags movies
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /movies [get]
func GetMovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movieCollection = database.OpenCollection(client, "movies")
		cursor, err := movieCollection.Find(ctx, bson.M{})
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

// @Summary Get movie by IMDb id
// @Tags movies
// @Produce json
// @Security ApiKeyAuth
// @Param imdbId path string true "IMDb ID"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /movie/{imdbId} [get]
func GetMovieByID(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		movieID := c.Param("imdbId")

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

// @Summary Search movies
// @Tags movies
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /searchmovies [get]
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

func GetRankings(client *mongo.Client, c *gin.Context) ([]models.Ranking, error) {
	var rankins []models.Ranking
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var rankingCollection *mongo.Collection = database.OpenCollection(client, "rankings")
	cursor, err := rankingCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &rankins); err != nil {
		return nil, err
	}

	return rankins, nil
}

// @Summary Get all genres
// @Tags movies
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /genres [get]
func GetGenres(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var genreCollection *mongo.Collection = database.OpenCollection(client, "genres")
		cursor, err := genreCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching genres"})
			return
		}
		defer cursor.Close(ctx)

		var genres []models.Genre
		if err = cursor.All(ctx, &genres); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while decoding genres"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": genres})
	}
}

func GetUsersFavouriteGenres(client *mongo.Client, userID string) ([]string, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}

	projection := bson.M{"favourite_movies_genres": 1, "_id": 0}

	options := options.FindOne().SetProjection(projection)

	var result struct {
		FavouriteMoviesGenres []models.Genre `bson:"favourite_movies_genres"`
	}

	var userCollection = database.OpenCollection(client, "users")
	err := userCollection.FindOne(ctx, filter, options).Decode(&result)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []string{}, nil
		}
		return nil, err
	}

	var favGenres []string
	for _, genre := range result.FavouriteMoviesGenres {
		favGenres = append(favGenres, genre.GenreName)
	}

	return favGenres, nil
}

func GetReviewRanking(admin_review string, client *mongo.Client, c *gin.Context) (string, int, error) {
	rankings, err := GetRankings(client, c)
	if err != nil {
		return "", 0, err
	}

	review_sentiment := ""

	for _, ranking := range rankings {
		if ranking.RankingValue != 999 { //999 is the neutral value
			review_sentiment = review_sentiment + ranking.RankingName + " "
		}
	}
	review_sentiment = strings.TrimSpace(review_sentiment) // Remove trailing space

	// Initialize OpenRouter LLM
	err = godotenv.Load(".env")
	if err != nil {
		return "", 0, fmt.Errorf("error loading .env file: %w", err)
	}
	openRouterApiKey := os.Getenv("OPENROUTER_API_KEY")
	if openRouterApiKey == "" {
		return "", 0, errors.New("OPENROUTER_API_KEY not set in .env file")
	}

	model := os.Getenv("OPENROUTER_MODEL_NAME")
	if model == "" {
		return "", 0, errors.New("OPENROUTER_MODEL_NAME not set in .env file")
	}

	llm, err := openai.New(
		openai.WithToken(openRouterApiKey),
		openai.WithBaseURL("https://openrouter.ai/api/v1"),
		openai.WithModel(model),
	)
	if err != nil {
		return "", 0, err
	}

	prompt_template := os.Getenv("BASE_PROMPT_TEMPLATE")
	if prompt_template == "" {
		return "", 0, errors.New("BASE_PROMPT_TEMPLATE not set in .env file")
	}

	prompt := strings.ReplaceAll(prompt_template, "{review_sentiment}", review_sentiment)
	prompt = strings.ReplaceAll(prompt, "{admin_review}", admin_review)

	requestCtx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	response, err := llm.Call(requestCtx, prompt)

	if err != nil {
		return "", 0, err
	}

	cleanResponse := strings.TrimSpace(response)
	if strings.EqualFold(cleanResponse, "Neutral") {
		return cleanResponse, 999, nil
	}

	rankvalue := 0
	for _, ranking := range rankings {
		if strings.EqualFold(cleanResponse, ranking.RankingName) {
			rankvalue = ranking.RankingValue
			break
		}
	}
	return cleanResponse, rankvalue, nil
}

// @Summary Update admin review for a movie
// @Tags movies
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param imdbId path string true "IMDb ID"
// @Param body body models.AdminReviewRequest true "Admin review"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /movie/review/{imdbId} [patch]
func UpdateAdminReview(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, err := utils.GetRoleFromCtx(c)
		if err != nil || role != "ADMIN" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		movieId := c.Param("imdbId")
		if movieId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Movie ID is required"})
			return
		}

		var req models.AdminReviewRequest
		var res models.AdminReviewResponse

		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sentiment, rankvalue, err := GetReviewRanking(req.AdminReview, client, c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while getting review ranking: " + err.Error()})
			return
		}

		filter := bson.M{"imdb_id": movieId}
		update := bson.M{
			"$set": bson.M{
				"admin_review": req.AdminReview,
				"ranking": bson.M{
					"ranking_name":  sentiment,
					"ranking_value": rankvalue,
				},
			},
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movieCollection = database.OpenCollection(client, "movies")
		result, err := movieCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while updating admin review"})
			return
		}
		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}

		res.RankingName = sentiment
		res.AdminReview = req.AdminReview

		c.JSON(http.StatusOK, gin.H{"data": res})
	}
}

// @Summary Get recommended movies
// @Tags movies
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /recommendatedmovies [get]
func GetMovieRecommendations(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := utils.GetuserIdFromCtx(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "useid not found in context"})
			return
		}
		favGenres, err := GetUsersFavouriteGenres(client, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching user's favourite genres"})
			return
		}

		err = godotenv.Load(".env")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error loading .env file"})
			return
		}

		var maxRecommendations int64 = 5
		limitStr := os.Getenv("RECOMMENDED_MOVIE_LIMIT")
		if limitStr != "" {
			var parsedLimit int
			_, err := fmt.Sscanf(limitStr, "%d", &parsedLimit)
			if err == nil && parsedLimit > 0 {
				maxRecommendations = int64(parsedLimit)
			}
		}

		findOption := options.Find()
		findOption.SetLimit(maxRecommendations)
		findOption.SetSort(bson.D{{Key: "ranking.ranking_value", Value: 1}}) // Sort by ranking value ascending

		filter := bson.M{"genres.genre_name": bson.M{"$in": favGenres}}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movieCollection = database.OpenCollection(client, "movies")
		cursor, err := movieCollection.Find(ctx, filter, findOption)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching recommended movies"})
			return
		}
		defer cursor.Close(ctx)

		recommendedMovies := []models.Movie{}
		if err = cursor.All(ctx, &recommendedMovies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while decoding recommended movies"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": recommendedMovies})
	}
}

// @Summary Get AI recommended movies
// @Tags movies
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /recommendations-ai [get]
func GetRecommendationFromAI(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := utils.GetuserIdFromCtx(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}

		favGenres, err := GetUsersFavouriteGenres(client, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching user's favourite genres"})
			return
		}

		if len(favGenres) == 0 {
			c.JSON(http.StatusOK, gin.H{"data": []models.Movie{}, "message": "No favorite genres found for user"})
			return
		}

		err = godotenv.Load(".env")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error loading .env file"})
			return
		}

		openRouterApiKey := os.Getenv("OPENROUTER_API_KEY")
		if openRouterApiKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "OPENROUTER_API_KEY not set"})
			return
		}

		model := os.Getenv("OPENROUTER_MODEL_NAME")
		if model == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "OPENROUTER_MODEL_NAME not set"})
			return
		}

		limitStr := os.Getenv("RECOMMENDED_MOVIE_LIMIT")
		limit := "5"
		if limitStr != "" {
			limit = limitStr
		}

		promptTemplate := os.Getenv("RECOMMENDATION_PROMPT_TEMPLATE")
		if promptTemplate == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "RECOMMENDATION_PROMPT_TEMPLATE not set"})
			return
		}

		genresStr := strings.Join(favGenres, ", ")
		prompt := strings.ReplaceAll(promptTemplate, "{genres}", genresStr)
		prompt = strings.ReplaceAll(prompt, "{limit}", limit)

		llm, err := openai.New(
			openai.WithToken(openRouterApiKey),
			openai.WithBaseURL("https://openrouter.ai/api/v1"),
			openai.WithModel(model),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error initializing AI model"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
		defer cancel()

		response, err := llm.Call(ctx, prompt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error calling AI model: " + err.Error()})
			return
		}

		cleanResponse := strings.TrimSpace(response)
		cleanResponse = strings.TrimPrefix(cleanResponse, "```json")
		cleanResponse = strings.TrimPrefix(cleanResponse, "```")
		cleanResponse = strings.TrimSuffix(cleanResponse, "```")
		cleanResponse = strings.TrimSpace(cleanResponse)

		var recommendedMovies []models.Movie
		if err := json.Unmarshal([]byte(cleanResponse), &recommendedMovies); err != nil {
			start := strings.Index(cleanResponse, "[")
			end := strings.LastIndex(cleanResponse, "]")
			if start != -1 && end != -1 && end > start {
				jsonPart := cleanResponse[start : end+1]
				if err := json.Unmarshal([]byte(jsonPart), &recommendedMovies); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing AI response: " + err.Error(), "raw_response": response})
					return
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing AI response: " + err.Error(), "raw_response": response})
				return
			}
		}

		for i := range recommendedMovies {
			recommendedMovies[i].ID = bson.NewObjectID()
		}

		c.JSON(http.StatusOK, gin.H{"data": recommendedMovies})
	}
}
