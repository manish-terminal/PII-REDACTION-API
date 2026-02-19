package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/asoasis/pii-redaction-api/internal/config"
	"github.com/asoasis/pii-redaction-api/internal/detector"
	"github.com/asoasis/pii-redaction-api/internal/handler"
	"github.com/asoasis/pii-redaction-api/internal/middleware"
	"github.com/asoasis/pii-redaction-api/internal/redactor"
	"github.com/asoasis/pii-redaction-api/internal/store"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	ctx := context.Background()
	dynamoStore, err := store.NewDynamoDBStore(ctx, cfg.AWSRegion, cfg.DynamoTableName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize DynamoDB store")
	}

	pipeline := detector.NewPipeline("en-US")
	redactorSvc := redactor.NewRedactor(dynamoStore)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}))

	r.Get("/v1/health", handler.Health)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg.APIKey))
		r.Post("/v1/redact", handler.NewRedactHandler(pipeline, redactorSvc).ServeHTTP)
		r.Post("/v1/detect", handler.NewDetectHandler(pipeline).ServeHTTP)
		r.Post("/v1/detokenize", handler.NewDetokenizeHandler(redactorSvc).ServeHTTP)
	})

	log.Info().Str("port", cfg.Port).Msg("Starting PII Redaction API")
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
