package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID                    bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID                string        `json:"user_id" bson:"user_id"`
	FirstName             string        `json:"first_name" bson:"first_name" validate:"required"`
	LastName              string        `json:"last_name" bson:"last_name" validate:"required"`
	Email                 string        `json:"email" bson:"email" validate:"required,email"`
	Password              string        `json:"password" bson:"password" validate:"required,min=8"`
	Role                  string        `json:"role" bson:"role" validate:"oneof=ADMIN USER GUEST"`
	CreatedAt             time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at" bson:"updated_at"`
	AccessToken           string        `json:"access_token" bson:"access_token"`
	RefreshToken          string        `json:"refresh_token" bson:"refresh_token"`
	FavouriteMoviesGenres []string      `json:"favourite_movies_genres" bson:"favourite_movies_genres" validate:"required,dive"`
}

type UserLogin struct {
	Email    string `json:"email" bson:"email" validate:"required,email"`
	Password string `json:"password" bson:"password" validate:"required,min=8"`
}

type LoginResponse struct {
	UserId                string   `json:"user_id"`
	FirstName             string   `json:"first_name"`
	LastName              string   `json:"last_name"`
	Email                 string   `json:"email"`
	Role                  string   `json:"role"`
	AccessToken           string   `json:"access_token"`
	RefreshToken          string   `json:"refresh_token"`
	FavouriteMoviesGenres []string `json:"favourite_movies_genres"`
}

type UserLogout struct {
	UserId string `json:"user_id"`
}

type UpdateUser struct {
	FirstName             *string   `json:"first_name,omitempty" bson:"first_name,omitempty"`
	LastName              *string   `json:"last_name,omitempty" bson:"last_name,omitempty"`
	Email                 *string   `json:"email,omitempty" bson:"email,omitempty"`
	Password              *string   `json:"password,omitempty" bson:"password,omitempty"`
	Role                  *string   `json:"role,omitempty" bson:"role,omitempty"`
	FavouriteMoviesGenres *[]string `json:"favourite_movies_genres,omitempty" bson:"favourite_movies_genres,omitempty"`
}
