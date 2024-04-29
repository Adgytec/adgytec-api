package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/rohan031/adgytec-api/version1/controllers"
)

func Router() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/newsletter", controllers.GetNewslettersEmail)

	return router
}
