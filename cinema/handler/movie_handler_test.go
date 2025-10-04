package handler

import (
	"bytes"
	"cinema/boxoffice"
	"cinema/model"
	"cinema/repository"
	"cinema/service"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type testMovieRepository struct {
	movies map[string]*model.Movie
}

func newTestMovieRepository() *testMovieRepository {
	return &testMovieRepository{movies: make(map[string]*model.Movie)}
}

func (r *testMovieRepository) Create(ctx context.Context, movie *model.Movie) error {
	key := strings.ToLower(movie.Title)
	if _, exists := r.movies[key]; exists {
		return repository.ErrMovieAlreadyExists
	}
	clone := *movie
	r.movies[key] = &clone
	return nil
}

func (r *testMovieRepository) UpdateSupplemental(ctx context.Context, movieID string, distributor *string, budget *int64, mpaRating *string, boxOffice *model.BoxOffice) error {
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

func (r *testMovieRepository) GetByTitle(ctx context.Context, title string) (*model.Movie, error) {
	if movie, ok := r.movies[strings.ToLower(title)]; ok {
		clone := *movie
		return &clone, nil
	}
	return nil, repository.ErrMovieNotFound
}

func (r *testMovieRepository) List(ctx context.Context, params repository.MovieListParams) ([]*model.Movie, error) {
	result := make([]*model.Movie, 0, len(r.movies))
	for _, movie := range r.movies {
		clone := *movie
		result = append(result, &clone)
	}
	return result, nil
}

type testBoxOfficeClient struct{}

func (testBoxOfficeClient) Fetch(ctx context.Context, title string) (*boxoffice.Record, error) {
	return nil, boxoffice.ErrNotFound
}

func TestCreateMovieHandlerReturnsCreated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newTestMovieRepository()
	svc := service.NewMovieService(repo, testBoxOfficeClient{})
	handler := NewMovieHandler(svc)

	payload := `{
        "title": "Test Movie 1",
        "genre": "Action",
        "releaseDate": "2023-01-15",
        "distributor": "Test Studios",
        "budget": 50000000,
        "mpaRating": "PG-13"
    }`

	req := httptest.NewRequest(http.MethodPost, "/movies", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateMovie(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusCreated, w.Code, w.Body.String())
	}
}
func TestCreateMovieHandlerAcceptsBOM(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newTestMovieRepository()
	svc := service.NewMovieService(repo, testBoxOfficeClient{})
	handler := NewMovieHandler(svc)

	basePayload := `{
        "title": "Another Test Movie",
        "genre": "Drama",
        "releaseDate": "2020-12-12",
        "distributor": "Some Studio",
        "budget": 1000000,
        "mpaRating": "PG"
    }`

	payload := append([]byte{0xEF, 0xBB, 0xBF}, []byte(basePayload)...)

	req := httptest.NewRequest(http.MethodPost, "/movies", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateMovie(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusCreated, w.Code, w.Body.String())
	}
}
