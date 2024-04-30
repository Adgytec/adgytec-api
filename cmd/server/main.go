package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rohan031/adgytec-api/helper"
	v1Router "github.com/rohan031/adgytec-api/v1/router"
)

type InvalidRequestError struct {
	message string
}

func (i *InvalidRequestError) Error() string {
	return i.message
}

func handle400(router *chi.Mux) {
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		var err = &InvalidRequestError{message: "404 route not found"}

		helper.ErrorResponse(w, err, http.StatusNotFound)
	})

	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		var err = &InvalidRequestError{message: "405 invalid request method"}

		helper.ErrorResponse(w, err, http.StatusMethodNotAllowed)
	})
}

func main() {
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
