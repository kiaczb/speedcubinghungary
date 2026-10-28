package router

import (
	"net/http"
	"speedcubinghungary/internal/allrounder"
	"speedcubinghungary/internal/auth"
	"speedcubinghungary/internal/certificate"
	"speedcubinghungary/internal/competition"
	"speedcubinghungary/internal/config"
	"speedcubinghungary/internal/middleware"
	"speedcubinghungary/internal/post"
	"speedcubinghungary/internal/prediction"
	"speedcubinghungary/internal/special"
	"speedcubinghungary/internal/stats"
	"speedcubinghungary/internal/types"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func New(
	cfg *config.Config,
	authMiddleware func(http.Handler) http.Handler,
	postHandler *post.Handler,
	competitionHandler *competition.Handler,
	predictionHandler *prediction.Handler,
	authHandler *auth.Handler,
	specialHandler *special.Handler,
	allrounderHandler *allrounder.Handler,
	statsHandler *stats.Handler,
	certificateHandler *certificate.Handler,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.SecurityHeaders)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	}))

	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir))))
	r.Get("/allrounder/stream", allrounderHandler.StreamUpdates)
	r.Get("/special/stream", specialHandler.StreamUpdates)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Use(middleware.RBACMiddleware(types.ROLE_ADMIN))
		r.Get("/wca/authorize", authHandler.AuthorizeWCA)
		r.Get("/certificates/assets/{id}", certificateHandler.GetAsset)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(cfg, middleware.LoginRateLimit))
		r.Get("/login", authHandler.Login)
		r.Get("/login/callback", authHandler.LoginCallback)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.JSON)

		r.Get("/posts", postHandler.GetPosts)
		r.Get("/posts/{id}", postHandler.GetPostByID)
		r.Get("/posts/latest", postHandler.GetLatestPosts)

		r.Get("/competitions/upcoming", competitionHandler.GetUpcoming)
		r.Get("/competitions/{id}/psych-sheet/{eventId}", competitionHandler.GetPsychSheet)
		r.Get("/competitions/{id}/competitors", competitionHandler.GetCompetitors)

		r.Get("/predictions", predictionHandler.GetContests)

		r.Get("/allrounder/standings", allrounderHandler.GetStandings)

		r.Get("/special/competitions", specialHandler.GetCompetitions)
		r.Get("/special/report", specialHandler.GetReport)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Get("/me", authHandler.Me)
			r.Post("/logout", authHandler.Logout)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Use(middleware.RBACMiddleware(types.ROLE_ADMIN))

			r.Post("/posts", postHandler.CreatePost)
			r.Put("/posts/{id}", postHandler.UpdatePost)
			r.Delete("/posts", postHandler.DeletePosts)

			r.Post("/predictions", predictionHandler.CreateContest)
			r.Delete("/predictions/{id}", predictionHandler.DeleteContest)

			r.Post("/allrounder/refresh", allrounderHandler.Refresh)

			r.Put("/special/results/{liveResultId}/attempts/{attemptNumber}/elapsed", specialHandler.RecordElapsed)

			r.Delete("/wca/grant", authHandler.RevokeWCA)

			r.Get("/stats/competitions", statsHandler.SearchCompetitions)
			r.Get("/stats/competitions/{id}", statsHandler.GetReport)

			r.Get("/certificates/recipients", certificateHandler.GetRecipients)
			r.Get("/certificates/templates", certificateHandler.ListTemplates)
			r.Post("/certificates/templates", certificateHandler.CreateTemplate)
			r.Get("/certificates/templates/{id}", certificateHandler.GetTemplate)
			r.Put("/certificates/templates/{id}", certificateHandler.UpdateTemplate)
			r.Delete("/certificates/templates/{id}", certificateHandler.DeleteTemplate)
			r.Get("/certificates/assets", certificateHandler.ListAssets)
			r.Post("/certificates/assets", certificateHandler.UploadAsset)
			r.Delete("/certificates/assets/{id}", certificateHandler.DeleteAsset)
		})
	})

	return r
}
