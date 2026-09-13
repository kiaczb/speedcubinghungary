// internal/router/router.go
package router

import (
	"net/http"
	"speedcubinghungary/internal/allrounder"
	"speedcubinghungary/internal/auth"
	"speedcubinghungary/internal/competition"
	"speedcubinghungary/internal/config"
	"speedcubinghungary/internal/middleware"
	"speedcubinghungary/internal/post"
	"speedcubinghungary/internal/prediction"
	"speedcubinghungary/internal/special"
	"speedcubinghungary/internal/types"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

// New összeállítja a HTTP routert. A wiring (repo → service → handler) a
// main.go-ban történik; ez a függvény kész handlereket vár és csak route-ol.
func New(
	cfg *config.Config,
	postHandler *post.Handler,
	competitionHandler *competition.Handler,
	predictionHandler *prediction.Handler,
	authHandler *auth.Handler,
	specialHandler *special.Handler,
	allrounderHandler *allrounder.Handler,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.JSON)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL}, // Use config
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true, // kell a cookie-hoz
	}))

	// Post routes
	r.Get("/posts", postHandler.GetPosts)
	r.Get("/posts/{id}", postHandler.GetPostByID)
	r.Get("/posts/latest", postHandler.GetLatestPosts)

	// Feltöltött posztképek kiszolgálása a volume-ról (publikus, csak olvasás).
	// A globális middleware.JSON minden válaszra application/json-t állít, és a Go
	// ServeContent NEM írja felül a már beállított Content-Type-ot — ezért itt
	// töröljük, hogy a fájlszerver a valódi képtípust tudja beállítani.
	fileServer := http.FileServer(http.Dir(cfg.UploadDir))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			w.Header().Del("Content-Type")
			fileServer.ServeHTTP(w, req)
		},
	)))

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddlewareWithConfig(cfg))
		r.Use(middleware.RBACMiddleware(types.ROLE_ADMIN))
		r.Post("/posts", postHandler.CreatePost)
		r.Put("/posts/{id}", postHandler.UpdatePost)
		r.Delete("/posts", postHandler.DeletePosts)
	})

	// Competition routes (WCA API proxy)
	r.Get("/competitions/upcoming", competitionHandler.GetUpcoming)

	r.Get("/predictions", predictionHandler.GetContests)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddlewareWithConfig(cfg))
		r.Use(middleware.RBACMiddleware(types.ROLE_ADMIN))
		r.Post("/predictions", predictionHandler.CreateContest)
	})

	r.Get("/allrounder/standings", allrounderHandler.GetStandings)
	r.Get("/allrounder/stream", allrounderHandler.StreamUpdates)

	r.Get("/special/competitions", specialHandler.GetCompetitions)
	r.Get("/special/report", specialHandler.GetReport)
	r.Get("/special/stream", specialHandler.StreamUpdates)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddlewareWithConfig(cfg))
		r.Use(middleware.RBACMiddleware(types.ROLE_ADMIN))

		r.Put("/special/results/{liveResultId}/attempts/{attemptNumber}/elapsed", specialHandler.RecordElapsed)
		r.Post("/allrounder/refresh", allrounderHandler.Refresh)
	})

	// Auth feature
	r.Get("/login/callback", authHandler.LoginCallback)

	// /me route - csak bejelentkezett user-eknek
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddlewareWithConfig(cfg)) // Use config
		r.Get("/me", authHandler.Me)
		r.Post("/logout", authHandler.Logout)
	})

	return r
}
