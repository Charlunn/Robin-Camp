CREATE TABLE IF NOT EXISTS movies (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL UNIQUE,
    genre TEXT NOT NULL,
    release_date DATE NOT NULL,
    distributor TEXT,
    budget BIGINT,
    mpa_rating TEXT,
    box_office JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_movies_created_at_id ON movies (created_at, id);
CREATE INDEX IF NOT EXISTS idx_movies_title_lower ON movies ((LOWER(title)));

CREATE TABLE IF NOT EXISTS ratings (
    movie_id UUID NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    rater_id TEXT NOT NULL,
    rating NUMERIC(2,1) NOT NULL CHECK (rating >= 0.5 AND rating <= 5.0 AND ((rating * 10) % 5 = 0)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (movie_id, rater_id)
);

CREATE INDEX IF NOT EXISTS idx_ratings_movie_id ON ratings (movie_id);
