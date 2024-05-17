package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// loading environment variables from .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	PORT := "8080"
	if port := os.Getenv("PORT"); port != "" {
		PORT = port
	}

	router, pool := initApp()
	defer pool.Close()

	log.Printf("Server is listening on PORT: %s", PORT)
	http.ListenAndServe(":"+PORT, router)
}

// TODO: add cursor (count) in user table for pagination
