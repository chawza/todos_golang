package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

var DB *sql.DB
var UserSession *SessionManager

func logRequest(r *http.Request) {
	InfoLogger.Println(r.Method, r.URL)
}

var UserRepo *UserRepostiory

func main() {
	InfoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	ErrorLogger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime)

	var err error

	DB, err = sql.Open("sqlite3", "./db.sqlite")

	if err != nil {
		panic(err)
	}

	defer DB.Close()

	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS tasks (title TEXT, description TEXT);")

	if err != nil {
		panic(err)
	}

	// templates
	homeTempl := template.Must(template.ParseFiles("templates/home.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "HMMMM")
	})

	taskRepo := TasksRepository{DB: DB}
	UserRepo = &UserRepostiory{DB: DB}
	UserSession = &SessionManager{DB: DB}

	err = UserSession.SetupSessionTable()

	if err != nil {
		panic(err)
	}

	// views
	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			homeTempl.Execute(
				w,
				struct{ Tasks []Task }{Tasks: taskRepo.fetchAll()},
			)
			logRequest(r)
			return

		} else if r.Method == http.MethodPost {
			taskRepo.add(r.FormValue("title"), r.FormValue("description"))
			http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
			logRequest(r)
			return
		}
	})

	http.HandleFunc("/tasks/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			index, _ := strconv.Atoi(r.PathValue("id"))
			err = taskRepo.deleteTaskById(index)

			if err != nil {
				panic(err)
			}

			http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
			logRequest(r)
		}
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		username := r.PathValue("username")
		password := r.PathValue("password")

		if len(username) == 0 || len(password) != 0 {
			// error status 400?
		}

		user_id, err := UserRepo.authenticate(username, password)

		if err != nil {
			panic(err)
		}

		user_session, err := UserSession.CreateSession(user_id)

		if err != nil {
			panic(err)
		}

		bodyPayload, err := json.Marshal(
			struct {
				Data LoginData `json:"data"`
			}{
				Data: LoginData{
					Token: user_session,
				},
			},
		)

		if err != nil {

		}

		w.Write(bodyPayload)
	})

	// fileServer := http.FileServer(http.Dir("static/"))
	// http.Handle("/static/", http.StripPrefix("/static/", fileServer))
	// http.ListenAndServe(":8000", nil)

	server := &http.Server{Addr: "127.0.0.1:8000", Handler: mux}
	InfoLogger.Println(fmt.Sprint("Listeningn in ", server.Addr))
	server.ListenAndServe()
}

type LoginData struct {
	Token string `json:"token"`
}
