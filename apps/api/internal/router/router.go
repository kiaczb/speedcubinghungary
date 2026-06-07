// internal/router/router.go
package router

import (
	"database/sql"
	"speedcubinghungary/internal/auth"
	"speedcubinghungary/internal/config"
	"speedcubinghungary/internal/middleware"
	"speedcubinghungary/internal/post"
	"speedcubinghungary/internal/types"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func New(db *sql.DB, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.JSON)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL}, // Use config
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true, // kell a cookie-hoz
	}))
	// Post routes
	repo := post.NewPostgresRepository(db)
	service := post.NewService(repo)
	handler := post.NewHandler(service)

	r.Get("/posts", handler.GetPosts)
	r.Get("/posts/{id}", handler.GetPostByID)
	r.Get("/posts/latest", handler.GetLatestPosts)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddlewareWithConfig(cfg))
		r.Use(middleware.RBACMiddleware(types.ROLE_ADMIN))
		r.Post("/posts", handler.CreatePost)
		r.Delete("/posts", handler.DeletePosts)
	})

	//Auth feature
	authRepo := auth.NewPostgresRepository(db)
	authService := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authService)

	r.Get("/login/callback", authHandler.LoginCallback)

	// /me route - csak bejelentkezett user-eknek
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddlewareWithConfig(cfg)) // Use config
		r.Get("/me", authHandler.Me)
		r.Post("/logout", authHandler.Logout)

	})

	return r
}
