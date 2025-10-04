package model

import "time"

type Rating struct {
    MovieID    string
    MovieTitle string
    RaterID    string
    Value      float64
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
