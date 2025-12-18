package utils

import (
	"context"
	"errors"
	"os"
	"time"

	"movie-app-go/database"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type SigninDetails struct {
	Email     string
	FirstName string
	LastName  string
	Role      string
	UserId    string
	jwt.RegisteredClaims
}

var JWT_SECRET_KEY = os.Getenv("JWT_SECRET_KEY")
var JWT_REFRESH_SECRET_KEY = os.Getenv("JWT_REFRESH_SECRET_KEY")

func GenerateAllTokens(userId, firstName, lastName, email, role string) (string, string, error) {
	claims := &SigninDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		UserId:    userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "movie-app-go",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(JWT_SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	refreshClaims := &SigninDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		UserId:    userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "movie-app-go",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(JWT_REFRESH_SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	return signedToken, signedRefreshToken, nil
}

func UpdateAllTokens(userId, token, refreshToken string, client *mongo.Client) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updateAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	data := bson.M{
		"$set": bson.M{
			"token":         token,
			"refresh_token": refreshToken,
			"updated_at":    updateAt,
		},
	}

	var userCollection *mongo.Collection = database.OpenCollection(client, "users")
	_, err = userCollection.UpdateOne(
		ctx,
		bson.M{"user_id": userId},
		data,
	)

	if err != nil {
		return err
	}
	return nil
}

func ValidateToken(tokenStr string) (*SigninDetails, error) {
	claims := &SigninDetails{}

	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(JWT_SECRET_KEY), nil
		},
	)

	if err != nil {
		return nil, err
	}

	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, err
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

func ValidateRefreshToken(tokenStr string) (*SigninDetails, error) {
	claims := &SigninDetails{}

	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(JWT_REFRESH_SECRET_KEY), nil
		},
	)

	if err != nil {
		return nil, err
	}

	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, err
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("refresh token expired")
	}

	return claims, nil
}

func GetAcessToken(c *gin.Context) (string, error) {
	if cookieToken, err := c.Cookie("access_token"); err == nil && cookieToken != "" {
		return cookieToken, nil
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("missing access token")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) > len(bearerPrefix) && authHeader[:len(bearerPrefix)] == bearerPrefix {
		return authHeader[len(bearerPrefix):], nil
	}

	return authHeader, nil
}

func GetuserIdFromCtx(c *gin.Context) (string, error) {
	userId, exists := c.Get("userId")

	if !exists {
		return "", errors.New("userId not found in context")
	}

	id, ok := userId.(string)

	if !ok {
		return "", errors.New("userId is not a string")
	}

	return id, nil
}

func GetRoleFromCtx(c *gin.Context) (string, error) {
	role, exists := c.Get("role")

	if !exists {
		return "", errors.New("role not found in context")
	}

	r, ok := role.(string)

	if !ok {
		return "", errors.New("role is not a string")
	}

	return r, nil
}
