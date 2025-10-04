package main

import (
	"cinema/boxoffice"
	"cinema/db"
	"cinema/handler"
	"cinema/handler/middleware"
	"cinema/repository"
	"cinema/service"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, falling back to environment variables")
	}

	port := getEnvOrDefault("PORT", "8080")
	authToken := os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		log.Fatal("AUTH_TOKEN must be provided for write operations")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be provided to connect to the database")
	}

	boxOfficeURL := os.Getenv("BOXOFFICE_URL")
	if boxOfficeURL == "" {
		log.Fatal("BOXOFFICE_URL must be provided for box office integration")
	}

	boxOfficeAPIKey := os.Getenv("BOXOFFICE_API_KEY")
	if boxOfficeAPIKey == "" {
		log.Fatal("BOXOFFICE_API_KEY must be provided for box office integration")
	}

	sqlDB, err := db.NewConnection(dbURL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer sqlDB.Close()

	movieRepo := repository.NewPostgresMovieRepository(sqlDB)
	ratingRepo := repository.NewPostgresRatingRepository(sqlDB)

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}
	boxOfficeClient := boxoffice.NewHTTPClient(boxOfficeURL, boxOfficeAPIKey, httpClient)

	movieService := service.NewMovieService(movieRepo, boxOfficeClient)
	ratingService := service.NewRatingService(movieRepo, ratingRepo)

	movieHandler := handler.NewMovieHandler(movieService)
	ratingHandler := handler.NewRatingHandler(ratingService)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.CORSMiddleware())

	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	authMiddleware := middleware.RequireBearerToken(authToken)

	router.GET("/movies", movieHandler.ListMovies)
	router.POST("/movies", authMiddleware, movieHandler.CreateMovie)
	router.POST("/movies/:title/ratings", ratingHandler.UpsertRating)
	router.GET("/movies/:title/rating", ratingHandler.GetAggregatedRating)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("server listening on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server stopped unexpectedly: %v", err)
	}
}

func getEnvOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
