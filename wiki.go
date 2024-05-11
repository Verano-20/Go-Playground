package main

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filePath := "data\\" + p.Title + ".txt"
	log.Printf("Writing to file '%v'.\n", filePath)
	return os.WriteFile(filePath, p.Body, 0600)
}

var templates = template.Must(template.ParseFiles("templates\\edit.html", "templates\\view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("invalid Page Title")
	}
	return m[2], nil
}

func loadPage(title string) (*Page, error) {
	log.Printf("loadPage: %v\n", title)
	filePath := "data\\" + title + ".txt"
	body, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(tmpl string, w http.ResponseWriter, p *Page) {
	log.Printf("renderTemplate: %v\n", tmpl)
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(string, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(m[2], w, r)
	}
}

func viewHandler(title string, w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling view request:\n%v\n\n", r)
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate("view", w, p)
}

func editHandler(title string, w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling edit request:\n%v\n\n", r)
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate("edit", w, p)
}

func saveHandler(title string, w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling save request:\n%v\n\n", r)
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Println("Http listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
