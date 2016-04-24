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

	"errors"

	"database/sql"

	"github.com/mahata/gwiki/util"
	_ "github.com/mattn/go-sqlite3"
	"github.com/russross/blackfriday"
)

type Config struct {
	TxtDir   string `json:"txtDir"`
	ImgDir   string `json:"imgDir`
	Password string `json:"password`
}

var templates = template.Must(template.ParseFiles("login.html", "edit.html", "view.html", "upload-file.html"))

var validPath = regexp.MustCompile(`^/(edit|save|view|static)/(\S+)$`)
var config Config

type Page struct {
	Title   string
	Content []byte
	Login   bool
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
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
	archiveDir := config.TxtDir + "/" + p.Title
	if !isExist(archiveDir) {
		err := os.Mkdir(archiveDir, 0700)
		if err != nil {
			os.Stderr.WriteString("Failed to create a directory. Is your disk full?")
			panic(err)
		}
	}
	archiveFilePath := archiveDir + "/" + string(fmt.Sprint(time.Now().Unix()))
	archiveErr := ioutil.WriteFile(archiveFilePath, p.Content, 0600)
	if archiveErr != nil {
		return archiveErr
	}
	filePath := archiveDir + ".txt"

	// FixMe: WIP
	db, err := sql.Open("sqlite3", "./sample.sqlite3")
	checkErr(err)
	stmt, err := db.Prepare("INSERT INTO wiki (title, content, unixtime) values(?, ?, ?)")
	checkErr(err)
	res, err := stmt.Exec(p.Title, p.Content, string(fmt.Sprint(time.Now().Unix())))
	checkErr(err)
	id, err := res.LastInsertId()
	checkErr(err)
	fmt.Println(id)

	return ioutil.WriteFile(filePath, p.Content, 0600)
}

func loadConf(confFile string) {
	file, err := ioutil.ReadFile(confFile)
	if err != nil {
		panic(err)
	}

	json.Unmarshal(file, &config)
}

func loadPage(title string) (*Page, error) {
	filename := config.TxtDir + "/" + title + ".txt"
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	//db, err := sql.Open("sqlite3", "./sample.sqlite3")
	//checkErr(err)
	//
	//rows, err := db.Query("SELECT * FROM wiki")
	//checkErr(err)
	//for rows.Next() {
	//	var id int
	//	var title string
	//	var content string
	//	var unixtime int
	//	err = rows.Scan(&id, &title, &content, &unixtime)
	//	fmt.Println(id, title, content, unixtime)
	//}

	return &Page{Title: title, Content: content}, nil
}

func Run(useNginx bool) {
	loadConf("config.json")

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/upload-file", uploadFileHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/static/", makeHandler(staticHandler))

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
	http.Redirect(w, r, "/view/index", http.StatusFound)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		p := &Page{
			Title:   "Login Page",
			Content: []byte("Login:"),
		}
		renderTemplate(w, "login", p)
	} else {
		if r.PostFormValue("password") == config.Password {
			cookie := http.Cookie{
				Name:     "login",
				Value:    toHash(config.Password),
				HttpOnly: true,
				Expires:  time.Now().AddDate(0, 1, 0),
			}
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			http.Error(w, "Login password doesn't match", http.StatusForbidden)
		}
	}
}

func calcExtension(contentType string) (string, error) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", nil
	case "image/gif":
		return ".gif", nil
	case "image/png":
		return ".png", nil
	default:
		return "", errors.New("Only jpeg, gif and png files are allowed for the file upload")
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

		extension, err := calcExtension(handler.Header.Get(("Content-Type")))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		uploadFileName := util.GetRandomString() + extension
		f, err := os.OpenFile(fmt.Sprintf("%s/%s", config.ImgDir, uploadFileName), os.O_WRONLY|os.O_CREATE, 0666)
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

		p := &Page{
			Title:   "File Posted",
			Content: []byte(fmt.Sprintf("![AltText](/static/%s \"TitleText\")", uploadFileName)),
		}
		renderTemplate(w, "upload-file", p)
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

	p.Content = blackfriday.MarkdownCommon([]byte(p.Content))
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		// Looks like there's no page to edit
		ioutil.WriteFile(fmt.Sprintf("%s/%s.txt", config.TxtDir, title), []byte(""), 0600)
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
	content := r.FormValue("content")
	p := &Page{Title: title, Content: []byte(content)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func staticHandler(w http.ResponseWriter, r *http.Request, fpath string) {
	imgPath := config.ImgDir + "/" + fpath
	_, err := os.Stat(imgPath)
	if err == nil {
		http.ServeFile(w, r, imgPath)
	} else {
		http.ServeFile(w, r, config.ImgDir+"/not-found.png") // FixMe: Add this image when installing the app
	}
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
