package routes

import (
	conntroller "movie-app-go/controllers"
	"movie-app-go/middleware"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupProtectedRoutes(router *gin.Engine, client *mongo.Client) {
	protectedRoutes := router.Group("/api/v1")
	protectedRoutes.Use(middleware.AuthenticationMiddleware())
	{
		protectedRoutes.GET("/getuserbyID/:userId", conntroller.GetUserByID(client))
		protectedRoutes.PUT("/updateuser/:userId", conntroller.UpdateUser(client))
		protectedRoutes.DELETE("/deleteuser/:userId", conntroller.DeleteUser(client))
		protectedRoutes.GET("/users", conntroller.GetUsers(client))
		protectedRoutes.POST("/addmovie", conntroller.AddMovie(client))
		protectedRoutes.GET("/movies", conntroller.GetMovies(client))
		protectedRoutes.GET("/movie/:imdbId", conntroller.GetMovieByID(client))
		protectedRoutes.PATCH("/movie/review/:imdbId", conntroller.UpdateAdminReview(client))
		protectedRoutes.GET("/recommendatedmovies", conntroller.GetMovieRecommendations(client))
		protectedRoutes.GET("/recommendations-ai", conntroller.GetRecommendationFromAI(client))
		protectedRoutes.GET("/searchmovies", conntroller.SearchMovies(client))
		protectedRoutes.GET("/genres", conntroller.GetGenres(client))
	}
}
