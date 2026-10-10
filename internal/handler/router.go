package handler

import (
	"net/http"
	"strings"
)

// NewRouter wires HTTP routes to handler methods.
func NewRouter(h *Handler, enableCORS bool) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", h.Health)
	mux.HandleFunc("/api/v1/resumes/analyze", method(http.MethodPost, h.UploadAndAnalyze))
	mux.HandleFunc("/api/v1/versions/generate", method(http.MethodPost, h.Generate))
	mux.HandleFunc("/api/v1/compare", method(http.MethodGet, h.Compare))

	mux.HandleFunc("/api/v1/resumes/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/structured") && r.Method == http.MethodPut:
			h.UpdateStructured(w, r)
		case strings.HasSuffix(path, "/analysis") && r.Method == http.MethodGet:
			h.GetAnalysisByResume(w, r)
		case r.Method == http.MethodGet:
			h.GetResume(w, r)
		case r.Method == http.MethodDelete:
			h.DeleteResume(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/api/v1/analyses/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/suggestions") && r.Method == http.MethodGet {
			h.ListSuggestions(w, r)
			return
		}
		http.NotFound(w, r)
	})

	mux.HandleFunc("/api/v1/suggestions/", method(http.MethodPatch, h.UpdateSuggestion))

	mux.HandleFunc("/api/v1/versions/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/export") && r.Method == http.MethodGet {
			h.Export(w, r)
			return
		}
		if r.Method == http.MethodGet {
			h.GetVersion(w, r)
			return
		}
		http.NotFound(w, r)
	})

	var handler http.Handler = mux
	if enableCORS {
		handler = corsMiddleware(handler)
	}
	return handler
}

func method(m string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != m {
			responseMethodNotAllowed(w)
			return
		}
		next(w, r)
	}
}

func responseMethodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusMethodNotAllowed)
	_, _ = w.Write([]byte(`{"code":405,"message":"method not allowed"}`))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
