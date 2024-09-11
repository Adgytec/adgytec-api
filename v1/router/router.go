package router

import (
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/rohan031/adgytec-api/v1/controllers"
	"github.com/rohan031/adgytec-api/v1/middleware"
)

func Router() *chi.Mux {
	router := chi.NewRouter()

	allowedOrigins := []string{
		"https://*.adgytec.in",
		"https://ecrimino.com",
		"https://jdkshipping.com",
	}
	if os.Getenv("ENV") == "dev" {
		allowedOrigins = append(allowedOrigins, "http://*")
	}

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	router.Get("/newsletter", controllers.GetNewslettersEmail)  // protected route called from dashboard to show all the emails that are signup for newsletter along with their status subscribe and unsubscribe
	router.Post("/newsletter", controllers.PostNewsletterEmail) // public route called from client frontend with their client token to add the email, if email already exists set status to subscribe
	// patch method for unsubscribing from email newsletter

	// user module
	router.Group(func(r chi.Router) {
		r.Use(middleware.TokenAuthentication)
		r.Use(middleware.UserRoleAuthorization)

		r.Post("/user", controllers.PostUser)
		r.Patch("/user/{id}", controllers.PatchUser)
		r.Delete("/user/{id}", controllers.DeleteUser)
		r.Get("/user/{id}", controllers.GetUserById)
		r.Get("/users", controllers.GetAllUsers)
	})

	// project module admin only routes
	router.Group(func(r chi.Router) {
		r.Use(middleware.TokenAuthentication)
		r.Use(middleware.AdminRoleAuthorization)

		r.Post("/project", controllers.PostProject)
		r.Post("/project/{projectId}/services", controllers.PostProjectAndServices)
		r.Post("/project/{projectId}/user", controllers.PostProjectAndUser)
		r.Get("/projects", controllers.GetAllProjects)
		r.Get("/project/{projectId}", controllers.GetProjectById)
		r.Get("/services", controllers.GetAllServices)
		r.Delete("/project/{projectId}", controllers.DeleteProjectById)
		r.Delete("/project/{projectId}/user", controllers.DeleteProjectAndUser)
		r.Delete("/project/{projectId}/services", controllers.DeleteProjectAndService)

		// project category management
		r.Post("/project/{projectId}/category", controllers.PostCategoryByProjectId)
		r.Patch("/project/{projectId}/category/{categoryId}", controllers.PatchCategoryById)
		r.Get("/project/{projectId}/category", controllers.GetCategoryByProjectId)
		r.Delete("/project/{projectId}/category/{categoryId}", controllers.DeleteCategoryById)

	})

	// project module users
	router.Group(func(r chi.Router) {
		r.Use(middleware.TokenAuthentication)

		r.Get("/client/projects", controllers.GetProjectsByUserId)
		r.Get("/client/projects/{projectId}/metadata", controllers.GetMetadataByProjectId)
	})

	// services
	// client token authentication for public endpoints
	router.Group(func(r chi.Router) {
		r.Use(middleware.ClientTokenAuthentication)
		// endpoints here

		r.Get("/services/news", controllers.GetAllNewsClient)

		// blogs
		r.Get("/services/blogs", controllers.GetAllBlogsByProjectIdClient)
		r.Get("/services/blog/{blogId}", controllers.GetBlogById)

		// gallery
		r.Get("/services/gallery/albums", controllers.GetAlbumsByProjectIdClient)
		// r.Get("/services/gallery/album/{albumId}")
	})

	// getting uuid
	router.Group(func(r chi.Router) {
		r.Use(middleware.TokenAuthentication)

		r.Get("/uuid", controllers.GetUUID)
	})

	//dashboard endpoints for services
	router.Group(func(r chi.Router) {
		r.Use(middleware.TokenAuthentication)
		r.Use(middleware.ServicesRoleAuthorization)

		// news
		r.Post("/services/news/{projectId}", controllers.PostNews)
		r.Get("/services/news/{projectId}", controllers.GetNews)
		r.Put("/services/news/{projectId}/{newsId}", controllers.PutNews)
		r.Delete("/services/news/{projectId}/{newsId}", controllers.DeleteNews)
		r.Delete("/services/news/{projectId}", controllers.DeleteNewsMultiple)

		// blogs
		r.Post("/services/blogs/{projectId}/{blogId}/media", controllers.PostMedia)
		r.Delete("/services/blogs/{projectId}/{blogId}/media", controllers.DeleteMedia)
		r.Post("/services/blogs/{projectId}/{blogId}", controllers.PostBlog)
		r.Get("/services/blogs/{projectId}", controllers.GetAllBlogsByProjectId)
		r.Get("/services/blogs/{projectId}/{blogId}", controllers.GetBlogById)
		r.Patch("/services/blogs/{projectId}/{blogId}", controllers.PatchBlogMetadataById)
		r.Delete("/services/blogs/{projectId}/{blogId}", controllers.DeleteBlogById)
		r.Patch("/services/blogs/{projectId}/{blogId}/cover", controllers.PatchBlogCover)
		r.Patch("/services/blogs/{projectId}/{blogId}/content", controllers.PatchBlogContent)

		// gallery
		r.Get("/services/gallery/{projectId}/albums", controllers.GetAlbumsByProjectId)
		r.Post("/services/gallery/{projectId}/albums", controllers.PostAlbum)
		r.Patch("/services/gallery/{projectId}/albums/{albumId}/metadata", controllers.PatchAlbumMetadataById)
		r.Patch("/services/gallery/{projectId}/albums/{albumId}/cover", controllers.PatchAlbumCoverById)
		r.Delete("/services/gallery/{projectId}/albums/{albumId}", controllers.DeleteAlbumById)
		r.Post("/services/gallery/{projectId}/album/{albumId}", controllers.PostPhoto)
		r.Get("/services/gallery/{projectId}/album/{albumId}", controllers.GetPhotosByAlbumId)
		r.Delete("/services/gallery/{projectId}/album/{albumId}", controllers.DeletePhotosById)
	})

	// public route
	router.Group(func(r chi.Router) {
		// middleware

		r.Post("/jdk/contact-us", controllers.PostContactUsJDK)
	})

	return router
}
