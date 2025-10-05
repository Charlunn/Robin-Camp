package repository

import (
	"cinema/internal/model"
	"context"
)

type RatingRepository interface {
	Upsert(ctx context.Context, rating *model.Rating) (bool, error)
	AggregateByMovieID(ctx context.Context, movieID string) (float64, int, error)
}
