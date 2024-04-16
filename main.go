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

type Contact struct {
	Email string
	Name  string
}

func newContact(email, name string) Contact {
	return Contact{
		Email: email,
		Name:  name,
	}
}

type AppState struct {
	Contacts []Contact
}

func newAppState() AppState {
	return AppState{
		Contacts: []Contact{
			newContact("diego@gmail.com", "diego"),
			newContact("bob@gmail.com", "bob"),
		},
	}
}

func main() {
	serverMux := http.NewServeMux()

	state := newAppState()

	t := newTemplate()

	serverMux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := t.Render(w, "index", state); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	serverMux.Handle("POST /contacts", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		name := r.FormValue("name")

		contact := newContact(email, name)

		state.Contacts = append(state.Contacts, contact)

		if err := t.Render(w, "contacts", state); err != nil {
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
