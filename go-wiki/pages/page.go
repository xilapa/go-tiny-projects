package pages

import (
	"html/template"
	"os"
	"regexp"
	"strings"
)

var regexInterPageLink = regexp.MustCompile(`\[([a-zA-Z0-9]+)\]`)

type Page struct {
	Title    string
	Body     []byte
	BodyView template.HTML
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

	sanit := template.HTMLEscapeString(string(body))

	bodyWithLinks := regexInterPageLink.ReplaceAllStringFunc(sanit,
		func(s string) string {
			match := regexInterPageLink.FindStringSubmatch(string(s))
			return `<a href="/view/` + match[1] + `">` + match[1] + `</a>`
		})
	bodyWithLinks = strings.Replace(bodyWithLinks, "\n", "<br>", -1)

	p := &Page{Title: title, Body: body, BodyView: template.HTML(bodyWithLinks)}

	return p, nil
}

func getFileName(title string) string {
	return "data/" + strings.ToLower(title) + ".txt"
}

func ParsePageTemplates() *template.Template {
	return template.Must(template.ParseFiles("pages/edit.html", "pages/view.html"))
}

func EnsureDataDirExists() error {
	return os.MkdirAll("data", os.ModePerm)
}
