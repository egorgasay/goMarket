package main

import (
	"fmt"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/httplog"
	echosession "github.com/go-session/echo-session"
	"github.com/labstack/echo"
	"gomarket/internal/logger"
	"gomarket/internal/market/config"
	handlers "gomarket/internal/market/handler"
	"gomarket/internal/market/storage"
	"gomarket/internal/market/usecase"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	cfg := config.New()

	log := httplog.NewLogger("market", httplog.Options{
		Concise: true,
	})

	repo, err := storage.Init(cfg.DBConfig)
	if err != nil {
		log.Info().Msg(fmt.Sprintf("Failed to initialize: %s", err.Error()))
	}

	logic := usecase.New(repo)
	e := echo.New()

	h := handlers.NewHandler(cfg, logic, logger.New(log))
	e.Use(echo.WrapMiddleware(httplog.RequestLogger(log)))
	e.Use(echo.WrapMiddleware(middleware.Recoverer))
	e.Use(echosession.New())

	t := &Template{
		templates: template.Must(template.ParseGlob("templates/html/*.html")),
	}
	e.Renderer = t

	h.PublicRoutes(e)
	h.PrivateRoutes(e)

	e.Static("/static", "static")

	// TODO: e.Use(middleware.Compress(4, ))
	go func() {
		log.Info().Msg("Stating market: " + cfg.Host)
		err := http.ListenAndServe(cfg.Host, e)
		if err != nil {
			log.Error().Msg(err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Info().Msg("Shutdown market ...")
}
