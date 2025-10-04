package model

import "time"

type Movie struct {
    ID           string
    Title        string
    Genre        string
    ReleaseDate  time.Time
    Distributor  *string
    Budget       *int64
    MpaRating    *string
    BoxOffice    *BoxOffice
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type BoxOffice struct {
    Revenue     BoxOfficeRevenue
    Currency    string
    Source      string
    LastUpdated time.Time
}

type BoxOfficeRevenue struct {
    Worldwide        int64
    OpeningWeekendUS *int64
}
