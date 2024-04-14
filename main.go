package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
)

type Template struct {
	tmpl *template.Template
}

func newTemplate() *Template {
	return &Template{
		tmpl: template.Must(template.ParseGlob("views/*.html")),
	}
}

func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	return t.tmpl.ExecuteTemplate(w, name, data)
}

type Count struct {
	Count int
}

func main() {
	serverMux := http.NewServeMux()

	count := Count{
		Count: 0,
	}

	t := newTemplate()

	serverMux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := t.Render(w, "index.html", count); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	serverMux.Handle("POST /count", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Count += 1
		if err := t.Render(w, "index.html", count); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	server := http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}

	fmt.Print("Server is running on port 8080")
	log.Fatal(server.ListenAndServe())

}
