package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type Todo struct {
	Title string
	Done  bool
}

type TodoPageData struct {
	PageTitle string
	Todos     []Todo
}

func main() {
	// template.Must 를 이용해 err 처리 생략
	// tmpl := template.Must(template.ParseFiles("layout.html"))

	tmpl, err := template.ParseFiles("layout.html")
	if err != nil {
		fmt.Println(err)
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := TodoPageData{
			PageTitle: "My TODO list",
			Todos: []Todo{
				{Title: "Task1", Done: false},
				{Title: "Task2", Done: true},
				{Title: "Task3", Done: true},
			},
		}

		tmpl.Execute(w, data)
	})
	http.ListenAndServe(":80", nil)
}
