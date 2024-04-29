package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rohan031/adgytec-api/version1/helper"
	v1Router "github.com/rohan031/adgytec-api/version1/router"
)

type InvalidRequestError struct {
	message string
}

func (i *InvalidRequestError) Error() string {
	return i.message
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

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		var err = &InvalidRequestError{message: "route not found"}

		helper.ErrorResponse(w, err, http.StatusNotFound)
	})

	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		var err = &InvalidRequestError{message: "method is not valid"}

		helper.ErrorResponse(w, err, http.StatusMethodNotAllowed)
	})

	log.Printf("Server is listening on PORT: %s", PORT)
	http.ListenAndServe(":"+PORT, router)
}
