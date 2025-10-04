package repository

import (
    "cinema/model"
    "context"
    "errors"
    "time"
)

var (
    ErrMovieNotFound      = errors.New("movie not found")
    ErrMovieAlreadyExists = errors.New("movie already exists")
)

type MovieCursor struct {
    CreatedAt time.Time
    ID        string
}

type MovieListParams struct {
    Q           string
    Year        *int
    Genre       *string
    Distributor *string
    BudgetLTE   *int64
    MpaRating   *string
    Limit       int
    After       *MovieCursor
}

type MovieRepository interface {
    Create(ctx context.Context, movie *model.Movie) error
    UpdateSupplemental(ctx context.Context, movieID string, distributor *string, budget *int64, mpaRating *string, boxOffice *model.BoxOffice) error
    GetByTitle(ctx context.Context, title string) (*model.Movie, error)
    List(ctx context.Context, params MovieListParams) ([]*model.Movie, error)
}
