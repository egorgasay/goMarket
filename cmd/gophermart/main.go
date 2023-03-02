package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	handlers "gomarket/internal/handler"
	"gomarket/internal/loyalty/config"
	"gomarket/internal/loyalty/storage"
	"gomarket/internal/usecase"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.New()

	storage, err := storage.Init(cfg.DBConfig)
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
		fmt.Println("Stating server: " + cfg.Host)
		log.Fatal(http.ListenAndServe(cfg.Host, router))
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutdown Server ...")
}
