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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"sigs.k8s.io/yaml"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, falling back to environment variables")
	}

	port := getEnvOrDefault("PORT", "8080")
	appEnv := strings.ToLower(getEnvOrDefault("APP_ENV", "production"))
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

	switch appEnv {
	case "development", "dev":
		gin.SetMode(gin.DebugMode)
	case "test", "testing":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.CORSMiddleware())

	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	router.GET("/swagger.json", func(c *gin.Context) {
		specPath := os.Getenv("OPENAPI_SPEC_PATH")
		if specPath == "" {
			specPath = "openapi.yml"
		}

		data, err := os.ReadFile(specPath)
		if err != nil {
			log.Printf("failed to read OpenAPI spec %s: %v", specPath, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "unable to load OpenAPI spec"})
			return
		}

		jsonData, err := yaml.YAMLToJSON(data)
		if err != nil {
			log.Printf("failed to convert OpenAPI spec to JSON: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "unable to parse OpenAPI spec"})
			return
		}

		c.Data(http.StatusOK, "application/json", jsonData)
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger.json")))

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

	log.Printf("server listening on port %s (env=%s)", port, appEnv)
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
