package main

import (
	"fmt"
	"net/http"
)

func handleFunc(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		fmt.Fprint(w, "<h1>hello, 欢迎来到goblog项目！</h1>")
	} else if r.URL.Path == "/about" {
		fmt.Fprint(w, "此博客用以记录编程笔记，如有反馈或建议，请联系"+
			"<a href=\"mailto:summer@example.com\">summer@example.com</a>")
	} else {
		fmt.Fprint(w, "<h1>请求页未找到 :(</h1>"+"<p>如有疑惑，请联系我们。</p>")
	}

}
func main() {
	http.HandleFunc("/", handleFunc)
	http.ListenAndServe("127.0.0.1:3000", nil)
}
