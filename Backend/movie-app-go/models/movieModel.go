package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Genre struct {
	GenreID   string `bson:"genre_id" json:"genre_id" validate:"required"`
	GenreName string `bson:"genre_name" json:"genre_name" validate:"required,min=2,max=100"`
}

type Ranking struct {
	RankingValue int    `bson:"ranking_value" json:"ranking_value" validate:"required"`
	RankingName  string `bson:"ranking_name" json:"ranking_name" validate:"required,min=2,max=100"`
}

type Movie struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ImdbID      string        `bson:"imdb_id" json:"imdb_id" validate:"required"`
	Title       string        `bson:"title" json:"title" validate:"required,min=2,max=200"`
	PosterURL   string        `bson:"poster_url" json:"poster_url" validate:"required,url"`
	YoutubeID   string        `bson:"youtube_id" json:"youtube_id" validate:"required"`
	Genres      []Genre       `bson:"genres" json:"genres" validate:"required,dive,required"`
	ReleaseYear int           `bson:"release_year" json:"release_year" validate:"required,min=1888,max=2100"`
	Ranking     Ranking       `bson:"ranking" json:"ranking" validate:"required"`
	AdminReview string        `bson:"admin_review" json:"admin_review" `
	Description string        `bson:"description" json:"description" validate:"required,min=10,max=5000"`
}

type AdminReviewRequest struct {
	AdminReview string `bson:"admin_review" json:"admin_review"`
}

type AdminReviewResponse struct {
	RankingName string `bson:"ranking_name" json:"ranking_name"`
	AdminReview string `bson:"admin_review" json:"admin_review"`
}
