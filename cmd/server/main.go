package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"stegno/internal/api"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	web := flag.String("web", "./web/dist", "built frontend directory (empty for API only)")
	maxUpload := flag.Int64("max-upload", 20<<20, "maximum multipart request bytes")
	flag.Parse()

	webDir := *web
	if webDir != "" {
		if info, err := os.Stat(filepath.Join(webDir, "index.html")); err != nil || info.IsDir() {
			log.Printf("frontend not found at %s; serving API only", webDir)
			webDir = ""
		}
	}
	srv := &http.Server{Addr: *addr, Handler: api.New(api.Config{WebDir: webDir, MaxUploadBytes: *maxUpload}), ReadHeaderTimeout: 10 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdown); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()
	log.Printf("listening on %s", *addr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
