package pages

import (
	"os"
	"text/template"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) Save() error {
	return os.WriteFile(getFileName(p.Title), p.Body, 0600)
}

func LoadPage(title string) (*Page, error) {
	filename := getFileName(title)
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func getFileName(title string) string {
	return "data/" + title + ".txt"
}

func ParsePageTemplates() *template.Template {
	return template.Must(template.ParseFiles("pages/edit.html", "pages/view.html"))
}

func EnsureDataDirExists() error {
	return os.MkdirAll("data", os.ModePerm)
}
