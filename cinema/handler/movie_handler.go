package handler

import (
    "cinema/model"
    "cinema/repository"
    "cinema/service"
    "errors"
    "net/http"
    "net/url"
    "strconv"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
)

type MovieHandler struct {
    service *service.MovieService
}

type createMovieRequest struct {
    Title       string  `json:"title"`
    Genre       string  `json:"genre"`
    ReleaseDate string  `json:"releaseDate"`
    Distributor *string `json:"distributor"`
    Budget      *int64  `json:"budget"`
    MpaRating   *string `json:"mpaRating"`
}

type movieResponse struct {
    ID          string             `json:"id"`
    Title       string             `json:"title"`
    Genre       string             `json:"genre"`
    ReleaseDate string             `json:"releaseDate"`
    Distributor *string            `json:"distributor,omitempty"`
    Budget      *int64             `json:"budget,omitempty"`
    MpaRating   *string            `json:"mpaRating,omitempty"`
    BoxOffice   *boxOfficeResponse `json:"boxOffice,omitempty"`
}

type boxOfficeResponse struct {
    Revenue     boxOfficeRevenueResponse `json:"revenue"`
    Currency    string                   `json:"currency"`
    Source      string                   `json:"source"`
    LastUpdated string                   `json:"lastUpdated"`
}

type boxOfficeRevenueResponse struct {
    Worldwide         int64  `json:"worldwide"`
    OpeningWeekendUSA *int64 `json:"openingWeekendUSA,omitempty"`
}

type moviePageResponse struct {
    Items      []movieResponse `json:"items"`
    NextCursor *string         `json:"nextCursor,omitempty"`
}

func NewMovieHandler(service *service.MovieService) *MovieHandler {
    return &MovieHandler{service: service}
}

func (h *MovieHandler) CreateMovie(c *gin.Context) {
    var req createMovieRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        writeError(c, http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY", "Malformed JSON payload", nil)
        return
    }

    if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Genre) == "" || strings.TrimSpace(req.ReleaseDate) == "" {
        writeError(c, http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY", "title, genre and releaseDate are required", nil)
        return
    }

    params := service.CreateMovieParams{
        Title:       req.Title,
        Genre:       req.Genre,
        ReleaseDate: req.ReleaseDate,
        Distributor: req.Distributor,
        Budget:      req.Budget,
        MpaRating:   req.MpaRating,
    }

    movie, err := h.service.CreateMovie(c.Request.Context(), params)
    switch {
    case err == nil:
        location := "/movies/" + url.PathEscape(movie.Title)
        c.Header("Location", location)
        c.JSON(http.StatusCreated, toMovieResponse(movie))
    case errors.Is(err, service.ErrInvalidInput):
        writeError(c, http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY", "Invalid request payload", nil)
    case errors.Is(err, repository.ErrMovieAlreadyExists):
        writeError(c, http.StatusConflict, "CONFLICT", "Movie with the same title already exists", nil)
    default:
        writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create movie", nil)
    }
}

func (h *MovieHandler) ListMovies(c *gin.Context) {
    var (
        yearParam       *int
        budgetParam     *int64
        genreParam      *string
        distributorParam *string
        mpaParam        *string
    )

    if value := strings.TrimSpace(c.Query("year")); value != "" {
        parsed, err := strconv.Atoi(value)
        if err != nil {
            writeError(c, http.StatusBadRequest, "BAD_REQUEST", "year must be an integer", nil)
            return
        }
        yearParam = &parsed
    }

    if value := strings.TrimSpace(c.Query("budget")); value != "" {
        parsed, err := strconv.ParseInt(value, 10, 64)
        if err != nil || parsed < 0 {
            writeError(c, http.StatusBadRequest, "BAD_REQUEST", "budget must be a non-negative integer", nil)
            return
        }
        budgetParam = &parsed
    }

    if value := strings.TrimSpace(c.Query("genre")); value != "" {
        genreParam = &value
    }

    if value := strings.TrimSpace(c.Query("distributor")); value != "" {
        distributorParam = &value
    }

    if value := strings.TrimSpace(c.Query("mpaRating")); value != "" {
        mpaParam = &value
    }

    limit := 0
    if value := strings.TrimSpace(c.Query("limit")); value != "" {
        parsed, err := strconv.Atoi(value)
        if err != nil {
            writeError(c, http.StatusBadRequest, "BAD_REQUEST", "limit must be an integer", nil)
            return
        }
        limit = parsed
    }

    params := service.ListMoviesParams{
        Q:           strings.TrimSpace(c.Query("q")),
        Year:        yearParam,
        Genre:       genreParam,
        Distributor: distributorParam,
        BudgetLTE:   budgetParam,
        MpaRating:   mpaParam,
        Limit:       limit,
        Cursor:      strings.TrimSpace(c.Query("cursor")),
    }

    movies, nextCursor, err := h.service.ListMovies(c.Request.Context(), params)
    switch {
    case err == nil:
        response := moviePageResponse{
            Items: make([]movieResponse, 0, len(movies)),
        }
        for _, movie := range movies {
            response.Items = append(response.Items, toMovieResponse(movie))
        }
        if nextCursor != nil {
            response.NextCursor = nextCursor
        }
        c.JSON(http.StatusOK, response)
    case errors.Is(err, service.ErrInvalidInput):
        writeError(c, http.StatusBadRequest, "BAD_REQUEST", "cursor is invalid", nil)
    default:
        writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list movies", nil)
    }
}

func toMovieResponse(movie *model.Movie) movieResponse {
    response := movieResponse{
        ID:          movie.ID,
        Title:       movie.Title,
        Genre:       movie.Genre,
        ReleaseDate: movie.ReleaseDate.Format("2006-01-02"),
        Distributor: movie.Distributor,
        Budget:      movie.Budget,
        MpaRating:   movie.MpaRating,
    }

    if movie.BoxOffice != nil {
        response.BoxOffice = &boxOfficeResponse{
            Revenue: boxOfficeRevenueResponse{
                Worldwide:         movie.BoxOffice.Revenue.Worldwide,
                OpeningWeekendUSA: movie.BoxOffice.Revenue.OpeningWeekendUS,
            },
            Currency: movie.BoxOffice.Currency,
            Source:   movie.BoxOffice.Source,
        }
        if !movie.BoxOffice.LastUpdated.IsZero() {
            response.BoxOffice.LastUpdated = movie.BoxOffice.LastUpdated.UTC().Format(time.RFC3339)
        }
    }

    return response
}
