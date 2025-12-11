package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"movie-app-go/database"
	"movie-app-go/models"
	"movie-app-go/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	HashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(HashedPassword), nil
}

func Signup(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User

		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validate := validator.New()
		if err := validate.Struct(user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		password, err := HashPassword(user.Password) // Hash the user's password

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while hashing password"})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second) // 100 seconds timeout
		defer cancel()                                                               // Ensure the context is cancelled to avoid memory leaks

		var userCollection = database.OpenCollection(client, "users")

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email}) // Check for existing user with the same email
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while checking for existing user"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
			return
		}

		user.UserID = bson.NewObjectID().Hex()
		user.Password = password

		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()

		data, err := userCollection.InsertOne(ctx, user) // Insert the new user into the database
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while creating user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": data})
	}
}

func Login(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userLogin models.UserLogin // Struct to hold login credentials

		if err := c.ShouldBindJSON(&userLogin); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var userCollection = database.OpenCollection(client, "users")
		var foundUser models.User

		err := userCollection.FindOne(ctx, bson.M{"email": userLogin.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(userLogin.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		accessToken, refreshToken, err := utils.GenerateAllTokens(foundUser.UserID, foundUser.FirstName, foundUser.LastName, foundUser.Email, foundUser.Role)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while generating tokens"})
			return
		}

		err = utils.UpdateAllTokens(foundUser.UserID, accessToken, refreshToken, client)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while updating tokens"})
			return
		}

		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "access_token",
			Value:    accessToken,
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
		})
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			Path:     "/",
			MaxAge:   7200,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
		})

		c.JSON(http.StatusOK, models.LoginResponse{
			FirstName:             foundUser.FirstName,
			LastName:              foundUser.LastName,
			Email:                 foundUser.Email,
			Role:                  foundUser.Role,
			AccessToken:           accessToken,
			RefreshToken:          refreshToken,
			FavouriteMoviesGenres: foundUser.FavouriteMoviesGenres,
		})
	}
}

func Logout(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var UserLogout models.UserLogout

		err := c.ShouldBindJSON(&UserLogout)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fmt.Println("User ID:", UserLogout.UserId)

		err = utils.UpdateAllTokens(UserLogout.UserId, "", "", client) // Clear tokens in the database

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while logging out"})
			return
		}

		// Clear cookies
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "access_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
		})
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
		})

		c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
	}
}

func GetUsers(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var userCollection = database.OpenCollection(client, "users")

		cursor, err := userCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching users"})
			return
		}
		defer cursor.Close(ctx)

		var users []models.User
		if err = cursor.All(ctx, &users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while decoding users"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": users})
	}
}

func GetUserByID(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var userCollection = database.OpenCollection(client, "users")

		var foundUser models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": foundUser})
	}
}

func UpdateUser(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		var user models.User

		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validate := validator.New() // Validate the user struct
		if err := validate.Struct(user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var userCollection = database.OpenCollection(client, "users")

		user.UpdatedAt = time.Now()

		update := bson.M{
			"$set": user,
		}

		_, err := userCollection.UpdateOne(ctx, bson.M{"user_id": userID}, update) // Update the user in the database
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while updating user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
	}
}

func DeleteUser(client *mongo.Client) gin.HandlerFunc {
	{
		return func(c *gin.Context) {
			userID := c.Param("id")

			var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
			defer cancel()

			var userCollection = database.OpenCollection(client, "users")

			_, err := userCollection.DeleteOne(ctx, bson.M{"user_id": userID}) // Delete the user from the database
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while deleting user"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
		}
	}
}

func RefreshToken(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second) // 100 seconds timeout
		defer cancel()                                                               // Ensure the context is cancelled to avoid memory leaks

		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token not found"})
			return
		}

		claims, err := utils.ValidateRefreshToken(refreshToken)
		if err != nil || claims == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}

		var userCollection = database.OpenCollection(client, "users")
		var foundUser models.User

		err = userCollection.FindOne(ctx, bson.M{"user_id": claims.UserId}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		newAccessToken, newRefreshToken, err := utils.GenerateAllTokens(foundUser.UserID, foundUser.FirstName, foundUser.LastName, foundUser.Email, foundUser.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while generating tokens"})
			return
		}

		err = utils.UpdateAllTokens(foundUser.UserID, newAccessToken, newRefreshToken, client)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while updating tokens"})
			return
		}

		c.SetCookie("access_token", newAccessToken, 86400, "/", "localhost", true, true)    // 1 day expiry
		c.SetCookie("refresh_token", newRefreshToken, 604800, "/", "localhost", true, true) // 7 days expiry

		c.JSON(http.StatusOK, gin.H{"message": "Tokens refreshed successfully"})
	}
}
