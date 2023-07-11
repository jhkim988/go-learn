package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/books/{title}/page/{page}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fmt.Fprint(w, vars["title"], vars["page"])
	})

	// http 서버의 메인 라우터에 대한 매개변수 r
	http.ListenAndServe(":80", r)
}
