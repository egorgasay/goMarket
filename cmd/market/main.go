package main

import (
	"github.com/go-chi/chi/middleware"
	"github.com/labstack/echo"
	"gomarket/internal/market/config"
	handlers "gomarket/internal/market/handler"
	"gomarket/internal/market/storage"
	"gomarket/internal/market/usecase"
	"html/template"
	"io"
	"log"
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

	repo, err := storage.Init(cfg.DBConfig)
	if err != nil {
		log.Fatalf("Failed to initialize: %s", err.Error())
	}

	logic := usecase.New(repo)
	e := echo.New()
	h := handlers.NewHandler(cfg, logic)
	e.Use(echo.WrapMiddleware(middleware.Logger))
	e.Use(echo.WrapMiddleware(middleware.Recoverer))

	t := &Template{
		templates: template.Must(template.ParseGlob("templates/html/*.html")),
	}
	e.Renderer = t

	e.GET("/", h.GetMain)
	e.Static("/static", "static")

	// TODO: router.Use(gzip.Gzip(gzip.BestSpeed))
	go func() {
		log.Println("Stating market: " + cfg.Host)
		log.Fatal(http.ListenAndServe(cfg.Host, e))
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutdown market ...")
}
