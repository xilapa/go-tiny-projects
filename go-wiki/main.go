package main

import (
	"log"
	"net/http"
	"regexp"

	"github.com/xilapa/go-tiny-projects/go-wiki/pages"
)

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := pages.LoadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := pages.LoadPage(title)
	if err != nil {
		p = &pages.Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/view/"+title, http.StatusFound)
		return
	}
	body := r.FormValue("body")
	p := &pages.Page{Title: title, Body: []byte(body)}
	if err := p.Save(); err != nil {
		writeError(w, err)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *pages.Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		writeError(w, err)
		return
	}
}

func writeError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

var templates = pages.ParsePageTemplates()
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func main() {
	if err := pages.EnsureDataDirExists(); err != nil {
		panic(err)
	}

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Println("wiki server started")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
