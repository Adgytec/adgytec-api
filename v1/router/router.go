package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/rohan031/adgytec-api/v1/controllers"
	"github.com/rohan031/adgytec-api/v1/middleware"
)

func Router() *chi.Mux {
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	router.Get("/newsletter", controllers.GetNewslettersEmail)  // protected route called from dashboard to showl all the emails that are signup for newsletter along with their status subscribe and unsbuscribe
	router.Post("/newsletter", controllers.PostNewsletterEmail) // public route called from client frontend with their client token to add the email, if email already exists set status to subscribe
	// patch method for unsubscribing from email newsletter

	// user module
	router.Group(func(r chi.Router) {
		r.Use(middleware.TokenAuthetication)
		r.Use(middleware.UserRoleAuthorization)

		r.Post("/user", controllers.PostUser)
		r.Patch("/user/{id}", controllers.PatchUser)
		r.Delete("/user/{id}", controllers.DeleteUser)
		r.Get("/user/{id}", controllers.GetUserById)
		r.Get("/users", controllers.GetAllUsers)
	})

	// project module
	router.Group(func(r chi.Router) {
		r.Use(middleware.TokenAuthetication)
		r.Use(middleware.AdminRoleAuthorization)

		r.Post("/project", controllers.PostProject)
		r.Post("/project/{projectId}/services", controllers.PostProjectAndServices)
		r.Post("/project/{projectId}/user", controllers.PostProjectAndUser)
	})

	// services
	// client token authentication for public endpoints
	router.Group(func(r chi.Router) {
		r.Use(middleware.ClientTokenAuthentication)
		// endpoints here

		r.Get("/services/news", controllers.GetAllNewsClient)
	})

	//dashboard endpoins for services
	router.Group(func(r chi.Router) {
		r.Use(middleware.TokenAuthetication)
		r.Use(middleware.ServicesRoleAuthorization)

		r.Post("/services/news/{projectId}", controllers.PostNews)
		r.Get("/services/news/{projectId}", controllers.GetNews)
		// r.Put("/services/news/{projectId}/{serviceId}")
		r.Delete("/services/news/{projectId}/{serviceId}", controllers.DeleteNews)
		r.Delete("/services/news/{projectId}", controllers.DeleteNewsMultiple)
	})

	return router
}
