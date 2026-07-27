package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/file-uploads-go/backend/internal/config"
	"github.com/file-uploads-go/backend/internal/server"
	"github.com/file-uploads-go/backend/pkg/upload"
	"github.com/file-uploads-go/backend/pkg/upload/storage"
)

func main() {
	cfg := config.Load()

	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatalf("create upload dir: %v", err)
	}

	store := storage.NewLocal(cfg.UploadDir)

	svc, err := upload.NewService(upload.Options{
		Config: upload.Config{
			UploadDir: cfg.UploadDir,
			MaxSize:   cfg.MaxSize,
		},
		Storage: store,
	})
	if err != nil {
		log.Fatalf("create upload service: %v", err)
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      server.New(svc, cfg.CORSOrigin).Router(),
		ReadTimeout:  30 * time.Minute,
		WriteTimeout: 30 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		log.Println("Shutting down server...")
		_ = srv.Shutdown(ctx)
	}()

	log.Printf("Upload service starting on port %s (dir=%s)", cfg.Port, cfg.UploadDir)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
