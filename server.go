package main

import (
	"database/sql"
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

	err = UserRepo.SetupTable()

	if err != nil {
		panic(err)
	}

	err = UserSession.SetupSessionTable()

	if err != nil {
		panic(err)
	}

	// UserRepo.DeleteById(1)

	// setup admin if not exist
	if UserRepo.CountAll() < 1 {
		PromptToCreateSuperUser()
	}

	// starts dead code
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
	// ends dead code

	authApi := AuthApi{}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/login", authApi.Login)

	server := &http.Server{Addr: "127.0.0.1:8000", Handler: mux}
	InfoLogger.Println(fmt.Sprint("Listening in ", server.Addr))
	server.ListenAndServe()
}
