package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/v1/database"
	v1Router "github.com/rohan031/adgytec-api/v1/router"
	"github.com/rohan031/adgytec-api/v1/services"
)

func handle400(router *chi.Mux) {
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		err := errors.New("404 route not found")

		helper.ErrorResponse(w, err, http.StatusNotFound)
	})

	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		err := errors.New("405 invalid request method")

		helper.ErrorResponse(w, err, http.StatusMethodNotAllowed)
	})
}

func main() {
	// loading environment variables from .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	// getting db connection pool
	pool, err := database.CreatePool()
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// setting database pool for use in services
	services.SetDatabasePool(pool)

	PORT := "8080"
	if port := os.Getenv("PORT"); port != "" {
		PORT = port
	}

	router := chi.NewRouter()

	// middleware
	router.Use(middleware.Heartbeat("/"))
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Mount("/v1", v1Router.Router())

	handle400(router)

	log.Printf("Server is listening on PORT: %s", PORT)
	http.ListenAndServe(":"+PORT, router)
}
