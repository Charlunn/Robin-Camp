package repository

import (
	"cinema/internal/model"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresMovieRepository struct {
	db *sql.DB
}

func NewPostgresMovieRepository(db *sql.DB) *PostgresMovieRepository {
	return &PostgresMovieRepository{db: db}
}

func (r *PostgresMovieRepository) Create(ctx context.Context, movie *model.Movie) error {
	const query = `
        INSERT INTO movies (id, title, genre, release_date, distributor, budget, mpa_rating, box_office)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	boxOfficeJSON, err := marshalBoxOffice(movie.BoxOffice)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(
		ctx,
		query,
		movie.ID,
		movie.Title,
		movie.Genre,
		movie.ReleaseDate,
		nullableString(movie.Distributor),
		nullableInt(movie.Budget),
		nullableString(movie.MpaRating),
		boxOfficeJSON,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrMovieAlreadyExists
		}
		return err
	}

	return nil
}

func (r *PostgresMovieRepository) UpdateSupplemental(ctx context.Context, movieID string, distributor *string, budget *int64, mpaRating *string, boxOffice *model.BoxOffice) error {
	const query = `
        UPDATE movies
        SET distributor = $2,
            budget = $3,
            mpa_rating = $4,
            box_office = $5,
            updated_at = NOW()
        WHERE id = $1
    `

	boxOfficeJSON, err := marshalBoxOffice(boxOffice)
	if err != nil {
		return err
	}

	res, err := r.db.ExecContext(
		ctx,
		query,
		movieID,
		nullableString(distributor),
		nullableInt(budget),
		nullableString(mpaRating),
		boxOfficeJSON,
	)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrMovieNotFound
	}
	return nil
}

func (r *PostgresMovieRepository) GetByTitle(ctx context.Context, title string) (*model.Movie, error) {
	const query = `
        SELECT id, title, genre, release_date, distributor, budget, mpa_rating, box_office, created_at, updated_at
        FROM movies
        WHERE LOWER(title) = LOWER($1)
    `

	var (
		movie        model.Movie
		distributor  sql.NullString
		budget       sql.NullInt64
		mpaRating    sql.NullString
		boxOfficeRaw []byte
	)

	err := r.db.QueryRowContext(ctx, query, title).Scan(
		&movie.ID,
		&movie.Title,
		&movie.Genre,
		&movie.ReleaseDate,
		&distributor,
		&budget,
		&mpaRating,
		&boxOfficeRaw,
		&movie.CreatedAt,
		&movie.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMovieNotFound
		}
		return nil, err
	}

	if distributor.Valid {
		movie.Distributor = &distributor.String
	}
	if budget.Valid {
		v := budget.Int64
		movie.Budget = &v
	}
	if mpaRating.Valid {
		movie.MpaRating = &mpaRating.String
	}
	if len(boxOfficeRaw) > 0 {
		boxOffice, err := unmarshalBoxOffice(boxOfficeRaw)
		if err != nil {
			return nil, err
		}
		movie.BoxOffice = boxOffice
	}

	return &movie, nil
}

func (r *PostgresMovieRepository) List(ctx context.Context, params MovieListParams) ([]*model.Movie, error) {
	base := strings.Builder{}
	base.WriteString(`
        SELECT id, title, genre, release_date, distributor, budget, mpa_rating, box_office, created_at, updated_at
        FROM movies
    `)

	var (
		clauses []string
		args    []interface{}
		idx     = 1
	)

	if params.Q != "" {
		clauses = append(clauses, fmt.Sprintf("title ILIKE '%%' || $%d || '%%'", idx))
		args = append(args, params.Q)
		idx++
	}

	if params.Year != nil {
		clauses = append(clauses, fmt.Sprintf("EXTRACT(YEAR FROM release_date) = $%d", idx))
		args = append(args, *params.Year)
		idx++
	}

	if params.Genre != nil && *params.Genre != "" {
		clauses = append(clauses, fmt.Sprintf("LOWER(genre) = LOWER($%d)", idx))
		args = append(args, *params.Genre)
		idx++
	}

	if params.Distributor != nil && *params.Distributor != "" {
		clauses = append(clauses, fmt.Sprintf("LOWER(distributor) = LOWER($%d)", idx))
		args = append(args, *params.Distributor)
		idx++
	}

	if params.BudgetLTE != nil {
		clauses = append(clauses, fmt.Sprintf("budget IS NOT NULL AND budget <= $%d", idx))
		args = append(args, *params.BudgetLTE)
		idx++
	}

	if params.MpaRating != nil && *params.MpaRating != "" {
		clauses = append(clauses, fmt.Sprintf("LOWER(mpa_rating) = LOWER($%d)", idx))
		args = append(args, *params.MpaRating)
		idx++
	}

	if params.After != nil {
		clauses = append(clauses, fmt.Sprintf("(created_at > $%d OR (created_at = $%d AND id > $%d))", idx, idx, idx+1))
		args = append(args, params.After.CreatedAt, params.After.ID)
		idx += 2
	}

	if len(clauses) > 0 {
		base.WriteString("WHERE ")
		base.WriteString(strings.Join(clauses, " AND "))
		base.WriteString("\n")
	}

	base.WriteString("ORDER BY created_at ASC, id ASC\n")
	base.WriteString(fmt.Sprintf("LIMIT $%d", idx))
	args = append(args, params.Limit)

	rows, err := r.db.QueryContext(ctx, base.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []*model.Movie
	for rows.Next() {
		var (
			movie        model.Movie
			distributor  sql.NullString
			budget       sql.NullInt64
			mpaRating    sql.NullString
			boxOfficeRaw []byte
		)

		if err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.Genre,
			&movie.ReleaseDate,
			&distributor,
			&budget,
			&mpaRating,
			&boxOfficeRaw,
			&movie.CreatedAt,
			&movie.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if distributor.Valid {
			movie.Distributor = &distributor.String
		}
		if budget.Valid {
			v := budget.Int64
			movie.Budget = &v
		}
		if mpaRating.Valid {
			movie.MpaRating = &mpaRating.String
		}
		if len(boxOfficeRaw) > 0 {
			boxOffice, err := unmarshalBoxOffice(boxOfficeRaw)
			if err != nil {
				return nil, err
			}
			movie.BoxOffice = boxOffice
		}

		movies = append(movies, &movie)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}

func marshalBoxOffice(boxOffice *model.BoxOffice) ([]byte, error) {
	if boxOffice == nil {
		return nil, nil
	}

	payload := struct {
		Revenue struct {
			Worldwide        int64  `json:"worldwide"`
			OpeningWeekendUS *int64 `json:"openingWeekendUSA"`
		} `json:"revenue"`
		Currency    string    `json:"currency"`
		Source      string    `json:"source"`
		LastUpdated time.Time `json:"lastUpdated"`
	}{
		Currency:    boxOffice.Currency,
		Source:      boxOffice.Source,
		LastUpdated: boxOffice.LastUpdated,
	}
	payload.Revenue.Worldwide = boxOffice.Revenue.Worldwide
	payload.Revenue.OpeningWeekendUS = boxOffice.Revenue.OpeningWeekendUS

	return json.Marshal(payload)
}

func unmarshalBoxOffice(raw []byte) (*model.BoxOffice, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var payload struct {
		Revenue struct {
			Worldwide        int64  `json:"worldwide"`
			OpeningWeekendUS *int64 `json:"openingWeekendUSA"`
		} `json:"revenue"`
		Currency    string    `json:"currency"`
		Source      string    `json:"source"`
		LastUpdated time.Time `json:"lastUpdated"`
	}

	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	return &model.BoxOffice{
		Revenue: model.BoxOfficeRevenue{
			Worldwide:        payload.Revenue.Worldwide,
			OpeningWeekendUS: payload.Revenue.OpeningWeekendUS,
		},
		Currency:    payload.Currency,
		Source:      payload.Source,
		LastUpdated: payload.LastUpdated,
	}, nil
}

func nullableString(value *string) interface{} {
	if value == nil {
		return nil
	}
	return *value
}

func nullableInt(value *int64) interface{} {
	if value == nil {
		return nil
	}
	return *value
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
