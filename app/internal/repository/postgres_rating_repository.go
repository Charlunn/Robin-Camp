package repository

import (
	"cinema/internal/model"
	"context"
	"database/sql"
)

type PostgresRatingRepository struct {
	db *sql.DB
}

func NewPostgresRatingRepository(db *sql.DB) *PostgresRatingRepository {
	return &PostgresRatingRepository{db: db}
}

func (r *PostgresRatingRepository) Upsert(ctx context.Context, rating *model.Rating) (bool, error) {
	const query = `
        INSERT INTO ratings (movie_id, rater_id, rating, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
        ON CONFLICT (movie_id, rater_id)
        DO UPDATE SET rating = EXCLUDED.rating, updated_at = NOW()
        RETURNING xmax = 0
    `

	var created bool
	if err := r.db.QueryRowContext(ctx, query, rating.MovieID, rating.RaterID, rating.Value).Scan(&created); err != nil {
		return false, err
	}
	return created, nil
}

func (r *PostgresRatingRepository) AggregateByMovieID(ctx context.Context, movieID string) (float64, int, error) {
	const query = `
        SELECT COALESCE(AVG(rating), 0), COUNT(*)
        FROM ratings
        WHERE movie_id = $1
    `

	var (
		average float64
		count   int
	)

	if err := r.db.QueryRowContext(ctx, query, movieID).Scan(&average, &count); err != nil {
		return 0, 0, err
	}

	return average, count, nil
}
