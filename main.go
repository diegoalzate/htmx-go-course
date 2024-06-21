package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
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

var id = 0

type Contact struct {
	Id    int
	Email string
	Name  string
}

func newContact(email, name string) Contact {
	id++
	return Contact{
		Id:    id,
		Email: email,
		Name:  name,
	}
}

type DBState struct {
	Contacts []Contact
}

func newDBState() DBState {
	return DBState{
		Contacts: []Contact{
			newContact("diego@gmail.com", "diego"),
			newContact("bob@gmail.com", "bob"),
		},
	}
}

func (db *DBState) hasEmail(val string) bool {
	found := false
	for _, c := range db.Contacts {
		if c.Email == val {
			found = true
		}
	}

	return found
}

func (db *DBState) indexOf(id int) (index int, found bool) {
	for idx, c := range db.Contacts {
		if c.Id == id {
			return idx, true
		}
	}

	return -1, false
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

type AppState struct {
	FormData FormData
}

func newAppState() AppState {
	return AppState{
		FormData: newFormData(),
	}
}

type State struct {
	App AppState
	DB  DBState
}

func newState() State {
	return State{
		App: newAppState(),
		DB:  newDBState(),
	}
}

func main() {
	serverMux := http.NewServeMux()

	state := newState()

	t := newTemplate()

	serverMux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := t.Render(w, "index", state); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	serverMux.HandleFunc("POST /contacts", func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		name := r.FormValue("name")

		if state.DB.hasEmail(email) {
			// validation error
			log.Print("db has email")
			state.App.FormData.Values["name"] = name
			state.App.FormData.Values["email"] = email
			state.App.FormData.Errors["email"] = "Email already exists"
			if err := t.Render(w, "create-contact-form", state.App.FormData); err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			return
		}

		// add event
		w.Header().Add("HX-Trigger", "newContact")

		contact := newContact(email, name)

		state.DB.Contacts = append(state.DB.Contacts, contact)

		err := t.Render(w, "create-contact-form", newFormData())

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	serverMux.HandleFunc("GET /contacts", func(w http.ResponseWriter, r *http.Request) {
		if err := t.Render(w, "contacts", state.DB); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	serverMux.HandleFunc("DELETE /contacts/{id}", func(w http.ResponseWriter, r *http.Request) {
		idPath := r.PathValue("id")
		id, err := strconv.Atoi(idPath)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		contactIdx, foundContact := state.DB.indexOf(id)

		if !foundContact {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		state.DB.Contacts = append(state.DB.Contacts[:contactIdx], state.DB.Contacts[contactIdx+1:]...)

	})

	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	server := http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}

	fmt.Print("Server is running on port 8080")
	log.Fatal(server.ListenAndServe())

}
