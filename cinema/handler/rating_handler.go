package handler

import (
	"cinema/repository"
	"cinema/service"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

const raterIDHeader = "X-Rater-Id"

type RatingHandler struct {
	service *service.RatingService
}

type upsertRatingRequest struct {
	Rating float64 `json:"rating"`
}

type ratingResponse struct {
	MovieTitle string  `json:"movieTitle"`
	RaterID    string  `json:"raterId"`
	Rating     float64 `json:"rating"`
}

type ratingAggregateResponse struct {
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}

func NewRatingHandler(service *service.RatingService) *RatingHandler {
	return &RatingHandler{service: service}
}

func (h *RatingHandler) UpsertRating(c *gin.Context) {
	title := c.Param("title")
	if strings.TrimSpace(title) == "" {
		writeError(c, http.StatusBadRequest, "BAD_REQUEST", "movie title is required", nil)
		return
	}

	raterID := strings.TrimSpace(c.GetHeader(raterIDHeader))
	if raterID == "" {
		writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid authentication information", nil)
		return
	}

	var req upsertRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY", "Malformed JSON payload", nil)
		return
	}

	rating, created, err := h.service.UpsertRating(c.Request.Context(), title, raterID, req.Rating)
	switch {
	case err == nil:
		status := http.StatusOK
		if created {
			status = http.StatusCreated
			location := fmt.Sprintf("/movies/%s/ratings/%s", url.PathEscape(rating.MovieTitle), url.PathEscape(rating.RaterID))
			c.Header("Location", location)
		}
		c.JSON(status, ratingResponse{
			MovieTitle: rating.MovieTitle,
			RaterID:    rating.RaterID,
			Rating:     rating.Value,
		})
	case errors.Is(err, service.ErrValidation):
		writeError(c, http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY", "rating must be between 0.5 and 5.0 in 0.5 steps", nil)
	case errors.Is(err, repository.ErrMovieNotFound):
		writeError(c, http.StatusNotFound, "NOT_FOUND", "Movie not found", nil)
	default:
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save rating", nil)
	}
}

func (h *RatingHandler) GetAggregatedRating(c *gin.Context) {
	title := c.Param("title")
	if strings.TrimSpace(title) == "" {
		writeError(c, http.StatusBadRequest, "BAD_REQUEST", "movie title is required", nil)
		return
	}

	average, count, err := h.service.GetAggregatedRating(c.Request.Context(), title)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, ratingAggregateResponse{Average: average, Count: count})
	case errors.Is(err, repository.ErrMovieNotFound):
		writeError(c, http.StatusNotFound, "NOT_FOUND", "Movie not found", nil)
	default:
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch rating aggregation", nil)
	}
}
