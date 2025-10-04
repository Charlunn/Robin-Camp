package service

import (
	"cinema/boxoffice"
	"cinema/model"
	"cinema/repository"
	"context"
	"strings"
	"testing"
)

type stubMovieRepository struct {
	movies map[string]*model.Movie
}

func newStubMovieRepository() *stubMovieRepository {
	return &stubMovieRepository{movies: make(map[string]*model.Movie)}
}

func (r *stubMovieRepository) Create(ctx context.Context, movie *model.Movie) error {
	key := strings.ToLower(movie.Title)
	if _, exists := r.movies[key]; exists {
		return repository.ErrMovieAlreadyExists
	}
	clone := *movie
	r.movies[key] = &clone
	return nil
}

func (r *stubMovieRepository) UpdateSupplemental(ctx context.Context, movieID string, distributor *string, budget *int64, mpaRating *string, boxOffice *model.BoxOffice) error {
	for _, movie := range r.movies {
		if movie.ID == movieID {
			movie.Distributor = distributor
			movie.Budget = budget
			movie.MpaRating = mpaRating
			movie.BoxOffice = boxOffice
			return nil
		}
	}
	return repository.ErrMovieNotFound
}

func (r *stubMovieRepository) GetByTitle(ctx context.Context, title string) (*model.Movie, error) {
	if movie, ok := r.movies[strings.ToLower(title)]; ok {
		clone := *movie
		return &clone, nil
	}
	return nil, repository.ErrMovieNotFound
}

func (r *stubMovieRepository) List(ctx context.Context, params repository.MovieListParams) ([]*model.Movie, error) {
	result := make([]*model.Movie, 0, len(r.movies))
	for _, movie := range r.movies {
		clone := *movie
		result = append(result, &clone)
	}
	return result, nil
}

type stubBoxOfficeClient struct{}

func (stubBoxOfficeClient) Fetch(ctx context.Context, title string) (*boxoffice.Record, error) {
	return nil, boxoffice.ErrNotFound
}

func TestCreateMovie_SucceedsWithValidInput(t *testing.T) {
	repo := newStubMovieRepository()
	svc := NewMovieService(repo, stubBoxOfficeClient{})

	distributor := "Test Studios"
	budget := int64(50000000)
	rating := "PG-13"

	params := CreateMovieParams{
		Title:       "Test Movie 1",
		Genre:       "Action",
		ReleaseDate: "2023-01-15",
		Distributor: &distributor,
		Budget:      &budget,
		MpaRating:   &rating,
	}

	movie, err := svc.CreateMovie(context.Background(), params)
	if err != nil {
		t.Fatalf("CreateMovie returned error: %v", err)
	}

	if movie.Title != params.Title {
		t.Fatalf("expected title %q, got %q", params.Title, movie.Title)
	}
}
