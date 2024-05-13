package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/rohan031/adgytec-api/v1/controllers"
)

func Router() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/newsletter", controllers.GetNewslettersEmail)  // protected route called from dashboard to showl all the emails that are signup for newsletter along with their status subscribe and unsbuscribe
	router.Post("/newsletter", controllers.PostNewsletterEmail) // public route called from client frontend with their client token to add the email, if email already exists set status to subscribe
	// patch method for unsubscribing from email newsletter

	router.Post("/user", controllers.PostUser)

	return router
}
