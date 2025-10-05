package handler

import (
	"strings"
	"testing"
)

func TestBindJSONBodySimple(t *testing.T) {
	json := `{"title":"MovieWin","releaseDate":"2023-02-01","genre":"Action"}`
	var req createMovieRequest
	if err := bindJSONBody(strings.NewReader(json), &req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Title != "MovieWin" {
		t.Fatalf("got %q", req.Title)
	}
}
