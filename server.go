package main

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	Rowid       int
	Title       string
	Description string
}

var taskList = []Task{}

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

var DB *sql.DB

func fetchAll() []Task {
	rows, err := DB.Query("SELECT rowid, title, description FROM tasks;")

	if err != nil {
		panic(err)
	}

	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var (
			rowid       int
			title       string
			description string
		)
		rows.Scan(&rowid, &title, &description)

		tasks = append(
			tasks,
			Task{
				Rowid:       rowid,
				Title:       title,
				Description: description,
			},
		)
	}

	return tasks
}

func getTaskbyId(rowid int) Task {
	rows, err := DB.Query(
		fmt.Sprintf(
			"SELECT title, description FROM tasks Where rowid=%d;",
			rowid,
		),
	)

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var (
			rowid       int
			title       string
			description string
		)
		rows.Scan(&rowid, &title, &description)

		rows.Close()
		return Task{
			Title:       title,
			Description: description,
		}
	}
	rows.Close()
	panic(fmt.Sprintf("Task with id %d not fonud", rowid))
}

func insertTask(title string, desc string) Task {
	result, err := DB.Exec(
		fmt.Sprintf(
			`INSERT INTO tasks (title, description) VALUES ("%s", "%s")`,
			title, desc,
		),
	)
	if err != nil {
		panic(err)
	}
	insertedId, err := result.LastInsertId()
	if err != nil {
		panic(err)
	}
	return getTaskbyId(int(insertedId))
}

func deleteTaskById(taskId int) error {
	result, err := DB.Exec(
		fmt.Sprintf(`DELETE FROM tasks WHERE rowid=%d`, taskId),
	)

	if err != nil {
		return err
	}

	affectedId, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affectedId == 0 {
		return errors.New("Not task effected! Trying to delete task with id " + strconv.Itoa(taskId))
	}

	return nil
}

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

	// views
	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			homeTempl.Execute(
				w,
				struct{ Tasks []Task }{Tasks: fetchAll()},
			)
			InfoLogger.Println(r.URL, w.Header().Get("status"))
			return

		} else if r.Method == http.MethodPost {
			insertTask(r.FormValue("title"), r.FormValue("description"))
			http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
			return
		}
	})

	http.HandleFunc("/tasks/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			index, _ := strconv.Atoi(r.PathValue("id"))
			err = deleteTaskById(index)

			if err != nil {
				panic(err)
			}

			http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
		}
	})

	fileServer := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	http.ListenAndServe(":8000", nil)
}
