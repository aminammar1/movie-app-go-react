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

// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body models.User true "User"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 409 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/register [post]
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

// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param body body models.UserLogin true "Credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/login [post]
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
			UserId:                foundUser.UserID,
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

// @Summary Logout
// @Tags auth
// @Accept json
// @Produce json
// @Param body body models.UserLogout true "Logout"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/logout [post]
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

// @Summary List users
// @Tags users
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/users [get]
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

// @Summary Get user by id
// @Tags users
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/getuserbyID/{userId} [get]
func GetUserByID(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userId")

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

// @Summary Update user
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "User ID"
// @Param body body models.UpdateUser true "Updates"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/updateuser/{userId} [put]
func UpdateUser(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userId")
		var user models.UpdateUser

		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		update := bson.M{}
		set := bson.M{}

		if user.FirstName != nil {
			set["first_name"] = *user.FirstName
		}
		if user.LastName != nil {
			set["last_name"] = *user.LastName
		}
		if user.Email != nil {
			set["email"] = *user.Email
		}
		if user.Password != nil {
			set["password"] = *user.Password
		}
		if user.Role != nil {
			set["role"] = *user.Role
		}
		if user.FavouriteMoviesGenres != nil {
			set["favourite_movies_genres"] = *user.FavouriteMoviesGenres
		}

		if len(set) > 0 {
			update["$set"] = set
		}

		if len(update) > 0 {
			var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
			defer cancel()

			var userCollection = database.OpenCollection(client, "users")

			_, err := userCollection.UpdateOne(ctx, bson.M{"user_id": userID}, update) // Update the user in the database
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while updating user"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
	}
}

// @Summary Delete user
// @Tags users
// @Produce json
// @Security ApiKeyAuth
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/deleteuser/{userId} [delete]
func DeleteUser(client *mongo.Client) gin.HandlerFunc {
	{
		return func(c *gin.Context) {
			userID := c.Param("userId")

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

// @Summary Refresh tokens
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/refresh-token [post]
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
