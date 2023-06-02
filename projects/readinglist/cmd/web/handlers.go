package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	books, err := app.readinglist.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	files := []string{
		"./ui/html/base.html",
		"./ui/html/partials/nav.html",
		"./ui/html/partials/header.html",
		"./ui/html/pages/home.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Error parsing files", http.StatusInternalServerError)
		return
	}

	err = ts.ExecuteTemplate(w, "base", books)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Error parsing files", http.StatusInternalServerError)
		return
	}
}

func (app *application) bookView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	book, err := app.readinglist.Get(int64(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	files := []string{
		"./ui/html/base.html",
		"./ui/html/partials/nav.html",
		"./ui/html/partials/header.html",
		"./ui/html/pages/view.html",
	}

	// used to convert comma-separated genres to a slice within the template
	funcs := template.FuncMap{"join": strings.Join}

	ts, err := template.New("showBook").Funcs(funcs).ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Error parsing files", http.StatusInternalServerError)
		return
	}

	err = ts.ExecuteTemplate(w, "base", book)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Error executing the template", http.StatusInternalServerError)
		return
	}
}

func (app *application) bookCreate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.bookCreateForm(w, r)
	case http.MethodPost:
		app.bookCreateProcess(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}

func (app *application) bookCreateForm(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.html",
		"./ui/html/partials/nav.html",
		"./ui/html/partials/header.html",
		"./ui/html/pages/create.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Error parsing files", http.StatusInternalServerError)
		return
	}

	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Error executing the template", http.StatusInternalServerError)
		return
	}
}

func (app *application) bookCreateProcess(w http.ResponseWriter, r *http.Request) {
	title := r.PostFormValue("title")
	if title == "" {
		http.Error(w, "error parsing title", http.StatusBadRequest)
		return
	}

	pages, err := strconv.Atoi(r.PostFormValue("pages"))
	if err != nil || pages < 1 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	published, err := strconv.Atoi(r.PostFormValue("published"))
	if err != nil || published < 1 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	genres := strings.Split(r.PostFormValue("genres"), " ")

	ratingFloat, err := strconv.ParseFloat(r.PostFormValue("rating"), 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rating := float32(ratingFloat)

	book := struct {
		Title     string   `json:"title"`
		Pages     int      `json:"pages"`
		Published int      `json:"published"`
		Genres    []string `json:"genres"`
		Rating    float32  `json:"rating"`
	}{
		Title:     title,
		Pages:     pages,
		Published: published,
		Genres:    genres,
		Rating:    rating,
	}

	data, err := json.Marshal(book)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("POST", app.readinglist.Endpoint, bytes.NewBuffer(data))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Printf("unexpected status %s", resp.Status)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
