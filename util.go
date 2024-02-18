package main

import (
	"fmt"
	"html/template"
	"log/slog"
	"net"
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
	t := template.New("").
		Funcs(template.FuncMap{
			"inc": func(i int) int {
				return i + 1
			},
		})

	return template.Must(
		t.ParseFS(
			TemplatesFS,
			fmt.Sprintf("templates/%s.html", name),
			"templates/base.html"),
	).Lookup("base.html")
}

func localIPAddress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		slog.Error(err.Error())
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return fmt.Sprintf("%s:%d", ipnet.IP, GlobalConfig.Port)
			}
		}
	}

	return ""
}
