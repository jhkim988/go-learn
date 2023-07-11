package main

import (
	"fmt"
	"net/http"
)

func main() {
	// 주소에 대해 핸들러 함수 등록
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, you've requesed: %s\n", r.URL.Path)
	})

	// 연결 수신 대기
	http.ListenAndServe(":80", nil)
}
