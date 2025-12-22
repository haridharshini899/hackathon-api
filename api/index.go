package handler

import (
	"hackathon-api/pkg/aggregator"
	"hackathon-api/pkg/api"
	"hackathon-api/pkg/fetcher"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

var handler http.Handler

func init() {
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
	apiHandler := api.NewHandler(aggService)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/api/v1/hackathons", apiHandler.GetHackathons)
	r.Get("/health", apiHandler.GetHealth)

	handler = r
}
func Handler(w http.ResponseWriter, r *http.Request) {
	handler.ServeHTTP(w, r)
}
