package service

import (
	"cinema/boxoffice"
	"cinema/model"
	"cinema/repository"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidInput = errors.New("invalid input")

type MovieService struct {
	repo            repository.MovieRepository
	boxOfficeClient boxoffice.Client
}

type CreateMovieParams struct {
	Title       string
	Genre       string
	ReleaseDate string
	Distributor *string
	Budget      *int64
	MpaRating   *string
}

type ListMoviesParams struct {
	Q           string
	Year        *int
	Genre       *string
	Distributor *string
	BudgetLTE   *int64
	MpaRating   *string
	Limit       int
	Cursor      string
}

func NewMovieService(repo repository.MovieRepository, client boxoffice.Client) *MovieService {
	return &MovieService{
		repo:            repo,
		boxOfficeClient: client,
	}
}

func (s *MovieService) CreateMovie(ctx context.Context, params CreateMovieParams) (*model.Movie, error) {
	title := strings.TrimSpace(params.Title)
	genre := strings.TrimSpace(params.Genre)
	if title == "" || genre == "" {
		return nil, ErrInvalidInput
	}

	releaseDate, err := time.Parse("2006-01-02", params.ReleaseDate)
	if err != nil {
		return nil, ErrInvalidInput
	}

	if params.Budget != nil && *params.Budget < 0 {
		return nil, ErrInvalidInput
	}

	movie := &model.Movie{
		ID:          uuid.NewString(),
		Title:       title,
		Genre:       genre,
		ReleaseDate: releaseDate,
		Distributor: params.Distributor,
		Budget:      params.Budget,
		MpaRating:   params.MpaRating,
	}

	if err := s.repo.Create(ctx, movie); err != nil {
		return nil, err
	}

	var (
		boxOffice *model.BoxOffice
		updated   bool
	)

	record, err := s.boxOfficeClient.Fetch(ctx, title)
	switch {
	case err == nil && record != nil:
		if movie.Distributor == nil && record.Distributor != nil {
			movie.Distributor = record.Distributor
			updated = true
		}
		if movie.Budget == nil && record.Budget != nil {
			movie.Budget = record.Budget
			updated = true
		}
		if movie.MpaRating == nil && record.MpaRating != nil {
			movie.MpaRating = record.MpaRating
			updated = true
		}

		boxOffice = &model.BoxOffice{
			Revenue: model.BoxOfficeRevenue{
				Worldwide:        record.Revenue.Worldwide,
				OpeningWeekendUS: record.Revenue.OpeningWeekendUS,
			},
			Currency:    record.Currency,
			Source:      record.Source,
			LastUpdated: record.LastUpdated,
		}
		updated = true
	case errors.Is(err, boxoffice.ErrNotFound):
		// graceful degradation: no box office data
	case err != nil:
		log.Printf("box office request failed (ignored for creation): %v", err)
	}

	if updated {
		if err := s.repo.UpdateSupplemental(ctx, movie.ID, movie.Distributor, movie.Budget, movie.MpaRating, boxOffice); err != nil {
			log.Printf("failed to update movie with box office data: %v", err)
		}
	}

	storedMovie, err := s.repo.GetByTitle(ctx, movie.Title)
	if err != nil {
		return nil, err
	}

	return storedMovie, nil
}

func (s *MovieService) ListMovies(ctx context.Context, params ListMoviesParams) ([]*model.Movie, *string, error) {
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	listParams := repository.MovieListParams{
		Q:           strings.TrimSpace(params.Q),
		Year:        params.Year,
		Genre:       params.Genre,
		Distributor: params.Distributor,
		BudgetLTE:   params.BudgetLTE,
		MpaRating:   params.MpaRating,
		Limit:       limit + 1,
	}

	if params.Cursor != "" {
		cursor, err := decodeCursor(params.Cursor)
		if err != nil {
			return nil, nil, ErrInvalidInput
		}
		listParams.After = cursor
	}

	movies, err := s.repo.List(ctx, listParams)
	if err != nil {
		return nil, nil, err
	}

	var nextCursor *string
	if len(movies) > limit {
		last := movies[len(movies)-1]
		movies = movies[:len(movies)-1]
		encoded, err := encodeCursor(last)
		if err != nil {
			return nil, nil, err
		}
		nextCursor = &encoded
	}

	return movies, nextCursor, nil
}

func encodeCursor(movie *model.Movie) (string, error) {
	payload := struct {
		CreatedAt time.Time `json:"createdAt"`
		ID        string    `json:"id"`
	}{
		CreatedAt: movie.CreatedAt,
		ID:        movie.ID,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(raw), nil
}

func decodeCursor(cursor string) (*repository.MovieCursor, error) {
	raw, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}

	var payload struct {
		CreatedAt time.Time `json:"createdAt"`
		ID        string    `json:"id"`
	}

	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	if payload.ID == "" || payload.CreatedAt.IsZero() {
		return nil, fmt.Errorf("cursor is missing required fields")
	}

	return &repository.MovieCursor{
		CreatedAt: payload.CreatedAt,
		ID:        payload.ID,
	}, nil
}
