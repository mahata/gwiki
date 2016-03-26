package wiki

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"regexp"
	"text/template"
	"time"

	"github.com/russross/blackfriday"
)

type Config struct {
	DataDir  string `json:"dataDir"`
	Password string `json:"password`
}

var templates = template.Must(template.ParseFiles("login.html", "edit.html", "view.html"))
var validPath = regexp.MustCompile("^/(login|edit|save|view)/([a-zA-Z0-9_-]+)$")
var config Config

type Page struct {
	Title string
	Body  []byte
	Login bool
}

func isExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func toHash(password string) string {
	converted := sha256.Sum256([]byte(password))
	return hex.EncodeToString(converted[:])
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

func Run(useNginx bool) {
	loadConf("config.json")

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/upload-file", uploadFileHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	if useNginx {
		listen, err := net.Listen("tcp", "0.0.0.0:8080")
		if err != nil {
			os.Stderr.WriteString("Failed to copy the HTTP server. Is port 8080 available?")
			panic(err)
		}
		fcgi.Serve(listen, nil)
	} else {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			os.Stderr.WriteString("Failed to copy the HTTP server. Is port 8080 available?")
			panic(err)
		}
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	p, _ := loadPage("index")
	p.Body = blackfriday.MarkdownCommon([]byte(p.Body))
	renderTemplate(w, "view", p)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		p := &Page{
			Title: "Login Page",
			Body:  []byte("Login:"),
		}
		renderTemplate(w, "login", p)
	} else {
		if r.PostFormValue("password") == config.Password {
			cookie := http.Cookie{
				Name:     "login",
				Value:    toHash(config.Password),
				HttpOnly: true,
				Expires:  time.Now().AddDate(0, 0, 1),
			}
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			http.Error(w, "Login password doesn't match", http.StatusForbidden)
		}
	}
}

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.Error(w, "Upload data have to be passed with POST.", http.StatusBadRequest)
	} else {
		r.ParseMultipartForm(32 << 20) // maxMemory
		file, handler, err := r.FormFile("upload-file")
		if err != nil {
			http.Error(w, "upload-file parameter must be passed", http.StatusBadRequest)
			return
		}
		defer file.Close()
		fmt.Fprintf(w, "%v", handler.Header) // To Be Deleted
		f, err := os.OpenFile("/tmp/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			http.Error(w, "Can't create a file handler. Disk full?", http.StatusInternalServerError)
			return
		}
		defer f.Close()
		_, err = io.Copy(f, file)
		if err != nil {
			http.Error(w, "Can't upload the upload-file. Disk full?", http.StatusInternalServerError)
			return
		}
	}
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

	cookie, err := r.Cookie("login")
	if err != nil {
		p.Login = false
	} else {
		p.Login = cookie.Value == toHash(config.Password)
	}

	p.Body = blackfriday.MarkdownCommon([]byte(p.Body))
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		// Looks like there's no page to edit
		ioutil.WriteFile(fmt.Sprintf("%s/%s.txt", config.DataDir, title), []byte(""), 0600)
		p, _ = loadPage(title)
	}

	cookie, err := r.Cookie("login")
	if err != nil {
		p.Login = false
	} else {
		p.Login = cookie.Value == toHash(config.Password)
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
