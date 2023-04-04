package main

import (
	"fmt"
	"net/http"
)

func handleFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>hello, 这是一个goblog项目！</h1>")
}
func main() {
	http.HandleFunc("/", handleFunc)
	http.ListenAndServe("127.0.0.1:3000", nil)
}
