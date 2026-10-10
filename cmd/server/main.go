package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cuixiangyu/ai-interview/internal/app"
	"github.com/cuixiangyu/ai-interview/internal/config"
)

func main() {
	cfg := config.Load()
	router := app.New(cfg)

	srv := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: cfg.RequestTimeout + 10*time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("AI Resume Optimizer listening on %s (ai=%s)", cfg.ServerAddr, cfg.AIProvider)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")
	_ = srv.Close()
}
