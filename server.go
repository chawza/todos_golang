package main

import (
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"strconv"
)

type Task struct {
	Title       string
	Description string
}

var taskList = []Task{}

func main() {
	homeTempl := template.Must(template.ParseFiles("templates/home.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "HMMMM")
	})

	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			homeTempl.Execute(
				w,
				struct{ Tasks []Task }{Tasks: taskList},
			)
			return

		} else if r.Method == http.MethodPost {
			detail := Task{
				Title:       r.FormValue("title"),
				Description: r.FormValue("description"),
			}

			taskList = append(taskList, detail)
			http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
			return
		}
	})

	http.HandleFunc("/tasks/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			index, _ := strconv.Atoi(r.PathValue("id"))
			taskList = slices.Delete(taskList, index, index+1)
			http.Redirect(w, r, "/home", http.StatusFound)
		}
	})

	fileServer := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	http.ListenAndServe(":8000", nil)
}
