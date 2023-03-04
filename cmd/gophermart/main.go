package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"gomarket/internal/logger"
	"gomarket/internal/loyalty/config"
	handlers "gomarket/internal/loyalty/handler"
	"gomarket/internal/loyalty/storage"
	"gomarket/internal/loyalty/usecase"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/httplog"
)

func main() {
	cfg := config.New()

	repo, err := storage.Init(cfg.DBConfig)
	if err != nil {
		log.Fatalf("Failed to initialize: %s", err.Error())
	}

	logic := usecase.New(repo)
	router := chi.NewRouter()

	log := httplog.NewLogger("loyalty", httplog.Options{
		Concise: true,
	})

	h := handlers.NewHandler(cfg, logic, logger.New(log))
	router.Use(httplog.RequestLogger(log))
	router.Use(middleware.Recoverer)

	router.Group(h.PublicRoutes)
	router.Group(h.PrivateRoutes)

	//router.Use(gzip.Gzip(gzip.BestSpeed))
	go func() {
		log.Info().Msg("Stating loyalty: " + cfg.Host)
		err := http.ListenAndServe(cfg.Host, router)
		if err != nil {
			log.Error().Msg(err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Info().Msg("Shutdown Server ...")
}
