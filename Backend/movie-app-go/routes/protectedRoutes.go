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
		protectedRoutes.GET("/user/:userId", conntroller.GetUserByID(client))
		protectedRoutes.PUT("/user/:userId", conntroller.UpdateUser(client))
		protectedRoutes.DELETE("/user/:userId", conntroller.DeleteUser(client))
		protectedRoutes.GET("/users", conntroller.GetUsers(client))
	}
}
