package view

import (
	"html/template"
	"net/http"
	"os"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("view",
	fx.Provide(NewHandler),
)

type Handler struct {
	template *template.Template
	logger   *zap.Logger
}

type HandlerParams struct {
	fx.In

	Logger *zap.Logger
}

func NewHandler(params HandlerParams) *Handler {
	wd, err := os.Getwd()
	if err != nil {
		params.Logger.Fatal("Failed to get working directory", zap.Error(err))
	}

	tmpl, err := template.ParseGlob(wd + "/internal/template/*.html")
	if err != nil {
		params.Logger.Fatal("Failed to parse templates", zap.Error(err))
	}

	return &Handler{
		template: tmpl,
	}
}

func (handler *Handler) RenderHomepage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := handler.template.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

func (handler *Handler) RenderCallPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := handler.template.ExecuteTemplate(w, "call.html", nil); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
