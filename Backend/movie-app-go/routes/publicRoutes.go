package routes

import (
	conntroller "movie-app-go/controllers"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"github.com/gin-gonic/gin"
)

func SetupPublicRoutes(router *gin.Engine, client *mongo.Client) {
	publicRoutes := router.Group("/api/v1")
	{
		publicRoutes.POST("/register", conntroller.Signup(client))
		publicRoutes.POST("/login", conntroller.Login(client))
		publicRoutes.POST("/refresh-token", conntroller.RefreshToken(client))
		publicRoutes.POST("/logout", conntroller.Logout(client))
	}
}