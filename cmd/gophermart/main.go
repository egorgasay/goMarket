package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"gomarket/config"
	handlers "gomarket/internal/handler"
	"gomarket/internal/repository"
	"gomarket/internal/usecase"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.New()

	storage, err := repository.New(cfg.DBConfig)
	if err != nil {
		log.Fatalf("Failed to initialize: %s", err.Error())
	}

	logic := usecase.New(storage)
	router := chi.NewRouter()
	h := handlers.NewHandler(cfg, logic)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Group(h.PublicRoutes)
	router.Group(h.PrivateRoutes)

	//router.Use(gzip.Gzip(gzip.BestSpeed))
	go func() {
		log.Fatal(http.ListenAndServe(cfg.Host, router))
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutdown Server ...")
}
