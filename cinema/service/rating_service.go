package service

import (
    "cinema/model"
    "cinema/repository"
    "context"
    "errors"
    "math"
)

var ErrValidation = errors.New("validation error")

type RatingService struct {
    movieRepo  repository.MovieRepository
    ratingRepo repository.RatingRepository
}

func NewRatingService(movieRepo repository.MovieRepository, ratingRepo repository.RatingRepository) *RatingService {
    return &RatingService{
        movieRepo:  movieRepo,
        ratingRepo: ratingRepo,
    }
}

func (s *RatingService) UpsertRating(ctx context.Context, movieTitle, raterID string, value float64) (*model.Rating, bool, error) {
    if !isValidRating(value) {
        return nil, false, ErrValidation
    }

    movie, err := s.movieRepo.GetByTitle(ctx, movieTitle)
    if err != nil {
        return nil, false, err
    }

    rating := &model.Rating{
        MovieID:    movie.ID,
        MovieTitle: movie.Title,
        RaterID:    raterID,
        Value:      value,
    }

    created, err := s.ratingRepo.Upsert(ctx, rating)
    if err != nil {
        return nil, false, err
    }

    return rating, created, nil
}

func (s *RatingService) GetAggregatedRating(ctx context.Context, movieTitle string) (float64, int, error) {
    movie, err := s.movieRepo.GetByTitle(ctx, movieTitle)
    if err != nil {
        return 0, 0, err
    }

    average, count, err := s.ratingRepo.AggregateByMovieID(ctx, movie.ID)
    if err != nil {
        return 0, 0, err
    }

    if count == 0 {
        return 0, 0, nil
    }

    rounded := math.Round(average*10) / 10
    return rounded, count, nil
}

func isValidRating(value float64) bool {
    if value < 0.5 || value > 5.0 {
        return false
    }
    scaled := value * 2
    return math.Abs(scaled-math.Round(scaled)) < 1e-9
}
