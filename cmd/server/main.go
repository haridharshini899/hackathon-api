package main

import (
	"fmt"
	"hackathon-api/pkg/aggregator"
	"hackathon-api/pkg/api"
	"hackathon-api/pkg/fetcher"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {

	fetchers := []fetcher.Fetcher{
		fetcher.NewDevfolioFetcher(),
		fetcher.NewUnstopFetcher(),
		fetcher.NewDevpostFetcher(),
		fetcher.NewMLHFetcher(),
		fetcher.NewHackerEarthFetcher(),
		fetcher.NewHack2SkillFetcher(),
		fetcher.NewReSkillFetcher(),
		fetcher.NewWhereUElevateFetcher(),
		fetcher.NewDevnovateFetcher(),
	}

	aggService := aggregator.NewService(fetchers)

	handler := api.NewHandler(aggService)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/api/v1/hackathons", handler.GetHackathons)
	r.Get("/health", handler.GetHealth)

	port := ":8080"
	fmt.Printf("Starting Hackathon Aggregator on %s...\n", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
