package main

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
)

type SlogMiddleware struct {
	Next http.Handler
}

func (s SlogMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Request", "method", r.Method, "url", r.URL.Path)
	s.Next.ServeHTTP(w, r)
}

func tmpl(name string) *template.Template {
	return template.Must(
		template.ParseFS(
			TemplatesFS,
			fmt.Sprintf("templates/%s.html", name),
			"templates/base.html"),
	).Lookup("base.html")
}
