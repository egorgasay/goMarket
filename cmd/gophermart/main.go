package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"gomarket/internal/loyalty/config"
	handlers "gomarket/internal/loyalty/handler"
	"gomarket/internal/loyalty/storage"
	"gomarket/internal/loyalty/usecase"
	middleware2 "gomarket/internal/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.New()

	repo, err := storage.Init(cfg.DBConfig)
	if err != nil {
		log.Fatalf("Failed to initialize: %s", err.Error())
	}

	logic := usecase.New(repo)
	router := chi.NewRouter()
	h := handlers.NewHandler(cfg, logic)

	logger := logrus.New()
	loggingMiddleware := middleware2.NewLoggingMiddleware(logger)
	router.Use(loggingMiddleware.Logging)
	router.Use(middleware.Recoverer)

	router.Group(h.PublicRoutes)
	router.Group(h.PrivateRoutes)

	//router.Use(gzip.Gzip(gzip.BestSpeed))
	go func() {
		fmt.Println("Stating server: " + cfg.Host)
		log.Fatal(http.ListenAndServe(cfg.Host, router))
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutdown Server ...")
}
