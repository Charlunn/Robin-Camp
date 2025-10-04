package boxoffice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrNotFound     = errors.New("box office record not found")
	ErrInvalidTitle = errors.New("movie title is required")
)

type Client interface {
	Fetch(ctx context.Context, title string) (*Record, error)
}

type HTTPClient struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
}

type Record struct {
	Distributor *string
	ReleaseDate string
	Budget      *int64
	MpaRating   *string
	Revenue     Revenue
	Currency    string
	Source      string
	LastUpdated time.Time
}

type Revenue struct {
	Worldwide        int64
	OpeningWeekendUS *int64
}

type apiResponse struct {
	Distributor *string        `json:"distributor"`
	ReleaseDate string         `json:"releaseDate"`
	Budget      *int64         `json:"budget"`
	Revenue     revenuePayload `json:"revenue"`
	MpaRating   *string        `json:"mpaRating"`
	Currency    string         `json:"currency"`
	Source      string         `json:"source"`
	LastUpdated string         `json:"lastUpdated"`
}

type revenuePayload struct {
	Worldwide         int64  `json:"worldwide"`
	OpeningWeekendUSA *int64 `json:"openingWeekendUSA"`
}

func NewHTTPClient(baseURL, apiKey string, httpClient *http.Client) Client {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		panic(fmt.Sprintf("invalid BOXOFFICE_URL: %v", err))
	}
	cloned := *parsed
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &HTTPClient{
		baseURL:    &cloned,
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

func (c *HTTPClient) Fetch(ctx context.Context, title string) (*Record, error) {
	if strings.TrimSpace(title) == "" {
		return nil, ErrInvalidTitle
	}

	endpoint := *c.baseURL
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/boxoffice"

	q := endpoint.Query()
	q.Set("title", title)
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("construct box office request failed: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call box office service failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var payload apiResponse
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return nil, fmt.Errorf("decode box office response failed: %w", err)
		}

		var lastUpdated time.Time
		if payload.LastUpdated != "" {
			parsed, err := time.Parse(time.RFC3339, payload.LastUpdated)
			if err != nil {
				return nil, fmt.Errorf("invalid lastUpdated format: %w", err)
			}
			lastUpdated = parsed
		}

		record := &Record{
			Distributor: payload.Distributor,
			ReleaseDate: payload.ReleaseDate,
			Budget:      payload.Budget,
			MpaRating:   payload.MpaRating,
			Revenue: Revenue{
				Worldwide:        payload.Revenue.Worldwide,
				OpeningWeekendUS: payload.Revenue.OpeningWeekendUSA,
			},
			Currency:    payload.Currency,
			Source:      payload.Source,
			LastUpdated: lastUpdated,
		}
		return record, nil
	case http.StatusNotFound:
		io.Copy(io.Discard, resp.Body)
		return nil, ErrNotFound
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		trimmed := strings.TrimSpace(string(body))
		if trimmed == "" {
			trimmed = resp.Status
		}
		return nil, fmt.Errorf("box office service returned status %d: %s", resp.StatusCode, trimmed)
	}
}
