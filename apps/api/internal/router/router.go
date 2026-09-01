// internal/router/router.go
package router

import (
	"database/sql"
	"net/http"
	"speedcubinghungary/internal/allrounder"
	"speedcubinghungary/internal/auth"
	"speedcubinghungary/internal/competition"
	"speedcubinghungary/internal/config"
	"speedcubinghungary/internal/middleware"
	"speedcubinghungary/internal/post"
	"speedcubinghungary/internal/special"
	"speedcubinghungary/internal/types"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

// New összeállítja a HTTP routert.
func New(
	db *sql.DB,
	cfg *config.Config,
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
	repo := post.NewPostgresRepository(db)
	service := post.NewService(repo)
	handler := post.NewHandler(service, cfg.UploadDir)

	r.Get("/posts", handler.GetPosts)
	r.Get("/posts/{id}", handler.GetPostByID)
	r.Get("/posts/latest", handler.GetLatestPosts)

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
		r.Post("/posts", handler.CreatePost)
		r.Put("/posts/{id}", handler.UpdatePost)
		r.Delete("/posts", handler.DeletePosts)
	})

	// Competition routes (WCA API proxy)
	competitionRepo := competition.NewWCARepository()
	competitionService := competition.NewService(competitionRepo)
	competitionHandler := competition.NewHandler(competitionService)

	r.Get("/competitions/upcoming", competitionHandler.GetUpcoming)

	r.Get("/allrounder/standings", allrounderHandler.GetStandings)
	r.Get("/allrounder/stream", allrounderHandler.StreamUpdates)

	// Budapest Special: WCA Live alapú eredménytábla a DNF-ek tényleges idejével.
	//
	// A versenyt a NÉZŐ választja ki (?competition=), nincs szerveroldali
	// "figyelt verseny". Az OLVASÁS publikus — ugyanez a verseny eredménytáblája
	// a WCA Live-on is bárki számára látható, a DNF-idők pedig épp azért
	// készülnek, hogy a versenyzők lássák őket. Egyedül a DNF-idő RÖGZÍTÉSE
	// admin, mert az a verseny üzemeltetése.
	r.Get("/special/competitions", specialHandler.GetCompetitions)
	r.Get("/special/report", specialHandler.GetReport)
	r.Get("/special/stream", specialHandler.StreamUpdates)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddlewareWithConfig(cfg))
		r.Use(middleware.RBACMiddleware(types.ROLE_ADMIN))

		r.Put("/special/results/{liveResultId}/attempts/{attemptNumber}/elapsed", specialHandler.RecordElapsed)
		r.Post("/allrounder/refresh", allrounderHandler.Refresh)
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
