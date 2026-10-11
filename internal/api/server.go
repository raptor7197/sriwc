// Package api serves the image watermarking HTTP API.
package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Config configures the optional frontend and multipart upload limit.
type Config struct {
	WebDir         string
	MaxUploadBytes int64
}

type server struct {
	maxUpload int64
	webDir    string
}

// New returns a handler with development CORS and all API routes.
func New(cfg Config) http.Handler {
	if cfg.MaxUploadBytes <= 0 {
		cfg.MaxUploadBytes = 20 << 20
	}
	s := &server{maxUpload: cfg.MaxUploadBytes, webDir: cfg.WebDir}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.URL.Path == "/api/health" {
			if r.Method != http.MethodGet {
				jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			jsonResponse(w, map[string]string{"status": "ok", "version": "1.0"})
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			var handler func(http.ResponseWriter, *http.Request)
			switch r.URL.Path {
			case "/api/embed":
				handler = s.embed
			case "/api/tamper":
				handler = s.tamper
			case "/api/verify":
				handler = s.verify
			case "/api/recover":
				handler = s.recoverImage
			case "/api/analyze":
				handler = s.analyze
			default:
				jsonError(w, http.StatusNotFound, "route not found")
				return
			}
			if r.Method != http.MethodPost {
				jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			defer func() {
				if r.MultipartForm != nil {
					_ = r.MultipartForm.RemoveAll()
				}
			}()
			handler(w, r)
			return
		}
		if r.Method != http.MethodGet {
			jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if s.webDir == "" {
			jsonError(w, http.StatusNotFound, "page not found")
			return
		}
		for _, component := range strings.Split(r.URL.Path, "/") {
			if component == ".." {
				jsonError(w, http.StatusNotFound, "page not found")
				return
			}
		}
		name := filepath.Join(s.webDir, filepath.FromSlash(strings.TrimPrefix(r.URL.Path, "/")))
		info, err := os.Stat(name)
		if err != nil || info.IsDir() && r.URL.Path != "/" {
			jsonError(w, http.StatusNotFound, "page not found")
			return
		}
		if r.URL.Path == "/" {
			name = filepath.Join(s.webDir, "index.html")
		}
		if info, err = os.Stat(name); err != nil || info.IsDir() {
			jsonError(w, http.StatusNotFound, "page not found")
			return
		}
		http.ServeFile(w, r, name)
	})
}

func jsonResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil { // A broken connection is no longer writable.
		_ = err
	}
}

func jsonError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (s *server) parse(w http.ResponseWriter, r *http.Request) bool {
	if r.ContentLength > s.maxUpload {
		jsonError(w, http.StatusRequestEntityTooLarge, "upload exceeds limit")
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, s.maxUpload)
	if err := r.ParseMultipartForm(4 << 20); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			jsonError(w, http.StatusRequestEntityTooLarge, "upload exceeds limit")
		} else {
			jsonError(w, http.StatusBadRequest, "invalid multipart form: "+err.Error())
		}
		return false
	}
	return true
}
