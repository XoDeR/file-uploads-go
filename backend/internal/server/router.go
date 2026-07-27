package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/file-uploads-go/backend/pkg/upload"
)

// Server wires HTTP handlers to the upload service.
type Server struct {
	svc        *upload.Service
	corsOrigin string
}

// New creates an HTTP server adapter.
func New(svc *upload.Service, corsOrigin string) *Server {
	return &Server{svc: svc, corsOrigin: corsOrigin}
}

// Router returns the chi router with all upload routes.
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(s.cors)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Route("/api/upload", func(r chi.Router) {
		r.Use(s.svc.RateLimiter().Middleware)

		r.Post("/stream", s.svc.HandleStream)

		cm := s.svc.ChunkedManager()
		r.Post("/init", cm.InitiateUpload)
		r.Post("/chunk", cm.UploadChunk)
		r.Post("/complete", cm.CompleteUpload)
		r.Get("/status", cm.GetUploadStatus)
		r.Get("/progress", s.svc.ProgressTracker().SSEHandler)
	})

	return r
}

func (s *Server) cors(next http.Handler) http.Handler {
	origin := s.corsOrigin
	if origin == "" {
		origin = "*"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Upload-ID, X-Upload-Type, X-Requested-With")
		w.Header().Set("Access-Control-Expose-Headers", "X-Upload-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
