package pages

import (
	"html/template"
	"os"
	"regexp"
	"strings"
	"time"

	lfucache "github.com/xilapa/go-tiny-projects/lfu-cache"
)

const maxCachePageCount = 2

var (
	regexInterPageLink = regexp.MustCompile(`\[([a-zA-Z0-9]+)\]`)
	pageCache          = lfucache.New(maxCachePageCount)
	hotPages           *Hotpages
)

type Page struct {
	Title    string
	Body     []byte
	BodyView template.HTML
}

func (p *Page) Save() error {
	return os.WriteFile(getFileName(p.Title), p.Body, 0600)
}

func LoadPage(title string) (*Page, error) {
	if page, ok := pageCache.Get(title); ok {
		return page.(*Page), nil
	}

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

	pageCache.Add(title, p)

	return p, nil
}

func getFileName(title string) string {
	return "data/" + strings.ToLower(title) + ".txt"
}

func ParsePageTemplates() *template.Template {
	return template.Must(template.ParseFiles("pages/edit.html", "pages/view.html", "pages/home.html"))
}

func EnsureDataDirExists() error {
	return os.MkdirAll("data", os.ModePerm)
}

type Hotpages struct {
	Names      []string
	Count      int
	expireDate time.Time
	full       bool
}

func LoadHotPages() *Hotpages {
	if hotPages != nil && hotPages.full && time.Now().After(hotPages.expireDate) {
		return hotPages
	}

	cacheCount := pageCache.Count()
	newHotPages := make([]string, cacheCount)
	i := 0
	for p := range pageCache.GetAllKeys() {
		newHotPages[i] = p
		i++
	}

	return &Hotpages{
		Names:      newHotPages,
		Count:      len(newHotPages),
		expireDate: time.Now().Add(time.Hour),
		full:       cacheCount == maxCachePageCount,
	}
}
