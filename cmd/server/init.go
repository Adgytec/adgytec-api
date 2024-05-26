package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rohan031/adgytec-api/database"
	"github.com/rohan031/adgytec-api/firebase"
	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/storage"
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

func initApp() (*chi.Mux, *pgxpool.Pool) {
	// init firebase
	firebaseClient, err := firebase.InitFirebaseAdminSdk()
	if err != nil {
		log.Fatal("Error connecting to firebase!!\n", err)
	}
	log.Println("Successfully connected to firebase!!")

	// init cloud storage
	minioClient, err := storage.InitCloudStorage()
	if err != nil {
		log.Fatal("Error creating minio-client!!\n", err)
	}
	log.Println("Successfully created mino storage client!!")

	// getting db connection pool
	pool, err := database.CreatePool()
	if err != nil {
		log.Fatal("Error connecting to database\n", err)
	}
	// setting database pool for use in services
	services.SetExternalConnection(pool, minioClient, firebaseClient)

	router := chi.NewRouter()

	// middleware
	router.Use(middleware.Heartbeat("/"))
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.AllowContentType("application/json", "multipart/form-data"))

	router.Mount("/v1", v1Router.Router())

	handle400(router)

	return router, pool
}
