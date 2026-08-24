package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"vulnkit/internal/api"
	"vulnkit/internal/db"
	"vulnkit/internal/docker"
	"vulnkit/internal/presets"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://vulnkit:vulnkit@localhost:5432/vulnkit?sslmode=disable"
	}

	database, err := db.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(ctx); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	presetStore := presets.NewStore(database.Pool)

	if err := presetStore.SeedDefaults(ctx); err != nil {
		log.Printf("Warning: seeding defaults failed: %v", err)
	}

	dockerClient, err := docker.NewClient()
	if err != nil {
		log.Printf("Warning: Docker not available: %v", err)
		dockerClient = &docker.Client{}
	}
	defer dockerClient.Close()

	projectDir := os.Getenv("PROJECT_DIR")
	if projectDir == "" {
		projectDir, _ = os.Getwd()
	}

	h := api.NewHandler(dockerClient, presetStore, projectDir)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:4173", "http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", h.Health)

		r.Route("/labs", func(r chi.Router) {
			r.Post("/generate", h.GenerateCompose)
			r.Post("/start", h.StartLab)
			r.Post("/stop", h.StopLab)
			r.Get("/status", h.GetStatus)
			r.Get("/logs/{id}", h.GetLogs)
			r.Get("/ports/check", h.CheckPort)

			r.Post("/containers/{id}/stop", h.StopContainer)
			r.Post("/containers/{id}/start", h.StartContainer)
			r.Delete("/containers/{id}", h.RemoveContainer)
		})

		r.Route("/presets", func(r chi.Router) {
			r.Get("/", h.ListPresets)
			r.Post("/", h.SavePreset)
			r.Delete("/{id}", h.DeletePreset)
		})

		r.Get("/hub/search", h.HubSearch)
		r.Get("/hub/tags", h.HubTags)

		r.Get("/labs/default", h.ListDefaultLabs)
		r.Get("/labs/default/status", h.DefaultLabStatus)
		r.Post("/labs/default/start", h.StartDefaultLab)
		r.Post("/labs/default/stop", h.StopDefaultLab)

		r.Get("/cve/search", h.CVESearch)
		r.Get("/cve/{id}", h.CVEDetail)

		r.Get("/labs/dynamic/check", h.DynamicBuildCheck)
		r.Post("/labs/dynamic/build", h.DynamicBuild)
		r.Get("/labs/dynamic/builds", h.ListDynamicBuilds)
		r.Get("/labs/dynamic/builds/{id}/dockerfile", h.GetBuildDockerfile)
		r.Delete("/labs/dynamic/builds/{id}", h.DeleteDynamicBuild)
	})

	fmt.Println("VulnKit backend running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
