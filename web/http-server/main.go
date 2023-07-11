package main

import (
	"fmt"
	"net/http"
)

func main() {
	// HTTP 서버
	// 동적 요청 처리
	http.HandleFunc("/http", func(w http.ResponseWriter, r *http.Request) {
		// r.URL.Query().Get("token") // Get 매개변수, Post 매개 변수 등을 읽을 수 있다.
		// r.FormValue("email")
		fmt.Fprint(w, "Welcome to my website/http")
	})

	// 정적 리소스 제공
	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 연결 수신 대기
	http.ListenAndServe(":80", nil)
}
