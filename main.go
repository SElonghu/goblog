package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"

	"github.com/gorilla/mux"
)

var db *sql.DB
var router = mux.NewRouter()

type Article struct {
	ID          int64
	Title, Body string
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>hello, 欢迎来到goblog项目！</h1>")
}
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "此博客用以记录编程笔记，如有反馈或建议，请联系"+
		"<a href=\"mailto:summer@example.com\">summer@example.com</a>")
}
func articlesShowHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	//fmt.Fprint(w, "文章id："+id)
	article := Article{}
	query := `SELECT * FROM articles WHERE id = ?`
	err := db.QueryRow(query, id).Scan(&article.ID, &article.Title, &article.Body)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 文章未找到")
		} else {
			checkError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500 服务器内部错误")
		}
	} else {
		//fmt.Fprint(w, "读取成功，标题为："+article.Title)
		templ, err := template.ParseFiles("./resources/views/articles/show.gohtml")
		checkError(err)
		err = templ.Execute(w, article)
		checkError(err)
	}
}
func articlesIndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "访问文章列表")
}

type ArticlesFormData struct {
	Title, Body string
	URL         *url.URL
	Errors      map[string]string
}

func articlesCreateHandler(w http.ResponseWriter, r *http.Request) {
	storeURL, _ := router.Get("articles.store").URL()
	data := ArticlesFormData{
		Title:  "",
		Body:   "",
		URL:    storeURL,
		Errors: nil,
	}
	tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		panic(err)
	}
}

func articlesStoreHandler(w http.ResponseWriter, r *http.Request) {

	title := r.PostFormValue("title")
	body := r.PostFormValue("body")
	errors := make(map[string]string)

	if title == "" {
		errors["title"] = "标题不能为空"
	} else if utf8.RuneCountInString(title) < 3 || utf8.RuneCountInString(title) > 40 {
		errors["title"] = "标题长度需介于3-40"
	}

	if body == "" {
		errors["body"] = "内容不能为空"
	} else if utf8.RuneCountInString(body) < 10 {
		errors["body"] = "内容不能少于10个字节"
	}

	if len(errors) == 0 {
		fmt.Fprint(w, "验证通过！")
		lastInsertID, err := saveArticleToDB(title, body)
		if lastInsertID > 0 {
			fmt.Fprint(w, "插入成功，ID为"+strconv.FormatInt(lastInsertID, 10))
		} else {
			checkError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500 服务器内部错误")
		}

	} else {
		fmt.Fprintf(w, "验证失败，errors: %v <br>", errors)
	}
	storeURL, _ := router.Get("articles.store").URL()
	data := ArticlesFormData{
		Title:  title,
		Body:   body,
		URL:    storeURL,
		Errors: errors,
	}

	//tmpl, err := template.New("create-form").Parse(html)
	tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		panic(err)
	}
}
func saveArticleToDB(title, body string) (int64, error) {
	var (
		err    error
		id     int64
		result sql.Result
		stmt   *sql.Stmt
	)
	stmt, err = db.Prepare("INSERT INTO articles (title, body) VALUES(?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	result, err = stmt.Exec(title, body)
	if err != nil {
		return 0, err
	}
	if id, err = result.LastInsertId(); id > 0 {
		return id, nil
	}
	return 0, err

}
func notFoundHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>请求页未找到 :(</h1>"+"<p>如有疑惑，请联系我们。</p>")
}
func forceHTMLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}
func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		next.ServeHTTP(w, r)
	})
}
func initDB() {
	var err error
	config := mysql.Config{
		User:                 "root",
		Passwd:               "Kylin*2020",
		Net:                  "tcp",
		Addr:                 "127.0.0.1",
		DBName:               "goblog",
		AllowNativePasswords: true,
	}
	db, err = sql.Open("mysql", config.FormatDSN())
	checkError(err)
	err = db.Ping()
	checkError(err)

}
func checkError(err error) {
	if err != nil {
		log.Fatal()
	}
}
func createTables() {
	createArticlesSQL := `CREATE TABLE IF NOT EXISTS articles(
		id bigint(20) PRIMARY KEY	AUTO_INCREMENT NOT NULL,
		title varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
		body longtext COLLATE utf8mb4_unicode_ci
	);`
	_, err := db.Exec(createArticlesSQL)
	checkError(err)
}
func main() {
	initDB()
	createTables()
	router.HandleFunc("/", homeHandler).Methods("GET").Name("home")
	router.HandleFunc("/about", aboutHandler).Methods("GET").Name("about")
	router.HandleFunc("/articles/{id:[1-9]+}", articlesShowHandler).Methods("GET").Name("articles.show")
	router.HandleFunc("/articles", articlesIndexHandler).Methods("GET").Name("articles.index")
	router.HandleFunc("/articles", articlesStoreHandler).Methods("POST").Name("articles.store")
	router.HandleFunc("/articles/create", articlesCreateHandler).Methods("GET").Name("articles.create")
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	router.Use(forceHTMLMiddleware)

	http.ListenAndServe("127.0.0.1:3000", removeTrailingSlash(router))
}
