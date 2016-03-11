package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"text/template"
	"time"

	"github.com/russross/blackfriday"
)

type Config struct {
	DataDir string `json:"dataDir"`
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var config Config

type Page struct {
	Title string
	Body  []byte
}

func isExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func (p *Page) save() error {
	archiveDir := config.DataDir + "/" + p.Title
	if !isExist(archiveDir) {
		err := os.Mkdir(archiveDir, 0700)
		if err != nil {
			os.Stderr.WriteString("Failed to create a directory. Is your disk full?")
			panic(err)
		}
	}
	archiveFilePath := archiveDir + "/" + string(fmt.Sprint(time.Now().Unix()))
	archiveErr := ioutil.WriteFile(archiveFilePath, p.Body, 0600)
	if archiveErr != nil {
		return archiveErr
	}
	filePath := archiveDir + ".txt"

	return ioutil.WriteFile(filePath, p.Body, 0600)
}

func loadConf(confFile string) {
	file, err := ioutil.ReadFile(confFile)
	if err != nil {
		panic(err)
	}

	json.Unmarshal(file, &config)
}

func loadPage(title string) (*Page, error) {
	filename := config.DataDir + "/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func main() {
	loadConf("config.json")

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		os.Stderr.WriteString("Failed to copy the HTTP server. Is port 8080 available?")
		panic(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	p, _ := loadPage("index")
	p.Body = blackfriday.MarkdownCommon([]byte(p.Body))
	renderTemplate(w, "view", p)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	p.Body = blackfriday.MarkdownCommon([]byte(p.Body))
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
