// web_app project main.go
package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Product struct {
	Id      int
	Model   string
	Company string
	Price   int
}

var database *sql.DB

func main() {

	db, err := sql.Open("mysql", "root:avalon1969@/productdb")

	if err != nil {
		log.Println(err)
	}
	database = db
	defer db.Close()

	router := mux.NewRouter()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	router.HandleFunc("/", IndexHandler)
	router.HandleFunc("/login", loginPage)

	router.HandleFunc("/logout", logoutPage)

	router.HandleFunc("/index", IndexHandler)
	router.HandleFunc("/create", CreateHandler)
	router.HandleFunc("/edit/{id:[0-9]+}", EditPage).Methods("GET")
	router.HandleFunc("/edit/{id:[0-9]+}", EditHandler).Methods("POST")
	router.HandleFunc("/delete/{id:[0-9]+}", DeleteHandler)

	http.Handle("/", router)

	fmt.Println("Server is listening...")
	http.ListenAndServe(":8000", nil)
}

func loggedOk(r *http.Request) bool {
	session, err := r.Cookie("session_id")
	loggedIn := (err != http.ErrNoCookie)
	if err == http.ErrNoCookie {
		return false
	}
	if loggedIn && session.Value == "neya1969" {
		return true
	}
	return false
}

func please_login(w http.ResponseWriter, r *http.Request) {
	if !loggedOk(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		//http.ServeFile(w, r, "templates/login.html")
		return
	}
}

func sessionData(methodName string, r *http.Request) {
	fmt.Println(methodName)
	session, err := r.Cookie("session_id")
	loggedIn := (err != http.ErrNoCookie)
	if err == http.ErrNoCookie {
		fmt.Println("err: ", err)
		fmt.Println("loggedIn: ", loggedIn)
		fmt.Println("___________________")
		return
	}
	fmt.Println("loggedIn: ", loggedIn)
	fmt.Println("func loggedOk: ", loggedOk(r))
	fmt.Println("session.Name: ", session.Name)
	fmt.Println("session.Value: ", session.Value)
	fmt.Println("session.Expires", session.Expires)
	fmt.Println("___________________")
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	sessionData("loginPage", r)
	login := "nologin"
	password := "nopassword"
	if r.Method == "POST" {

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		login = r.FormValue("login")
		password = r.FormValue("password")

		if login != "neya1969" ||
			password != "avalon1969" {
			fmt.Println("!!! err autorize")
			http.ServeFile(w, r, "templates/login.html")
		}

		//http.Redirect(w, r, "/", 301)
	} else {
		http.ServeFile(w, r, "templates/login.html")
	}

	expiration := time.Now().Add(10 * time.Hour)
	cookie := http.Cookie{
		Name:    "session_id",
		Value:   login,
		Expires: expiration,
	}
	http.SetCookie(w, &cookie)
	//sessionData("# loginPage", r)
	http.Redirect(w, r, "/", http.StatusFound)
}

func logoutPage(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		http.ServeFile(w, r, "templates/login.html")
		return
	}
	session.Expires = time.Now().AddDate(0, 0, -1)
	//session.Value = "nologin"
	http.SetCookie(w, session)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {

	sessionData("IndexHandler", r)

	rows, err := database.Query("select * from productdb.Products")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	products := []Product{}

	for rows.Next() {
		p := Product{}
		err := rows.Scan(&p.Id, &p.Model, &p.Company, &p.Price)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
	}

	tmpl, _ := template.ParseFiles("templates/index.html")
	tmpl.Execute(w, products)
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	//please_login(w, r)
	if r.Method == "POST" {

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		model := r.FormValue("model")
		company := r.FormValue("company")
		price := r.FormValue("price")

		_, err = database.Exec("insert into productdb.Products (model, company, price) values (?, ?, ?)",
			model, company, price)

		if err != nil {
			log.Println(err)
		}
		http.Redirect(w, r, "/", 301)
	} else {
		http.ServeFile(w, r, "templates/create.html")
	}
}

// возвращаем пользователю страницу для редактирования объекта
func EditPage(w http.ResponseWriter, r *http.Request) {
	//please_login(w, r)
	vars := mux.Vars(r)
	id := vars["id"]

	row := database.QueryRow("select * from productdb.Products where id = ?", id)
	prod := Product{}
	err := row.Scan(&prod.Id, &prod.Model, &prod.Company, &prod.Price)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(404), http.StatusNotFound)
	} else {
		tmpl, _ := template.ParseFiles("templates/edit.html")
		tmpl.Execute(w, prod)
	}
}

// получаем измененные данные и сохраняем их в БД
func EditHandler(w http.ResponseWriter, r *http.Request) {
	//please_login(w, r)
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	id := r.FormValue("id")
	model := r.FormValue("model")
	company := r.FormValue("company")
	price := r.FormValue("price")

	_, err = database.Exec("update productdb.Products set model=?, company=?, price = ? where id = ?",
		model, company, price, id)

	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/", 301)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	//please_login(w, r)
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := database.Exec("delete from productdb.Products where id = ?", id)
	if err != nil {
		log.Println(err)
	}

	http.Redirect(w, r, "/", 301)
}
