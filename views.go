package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"text/template"
)

type SiteResponse struct {
	Username string
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("favicon handler")
}

type Context struct {
	P *Persistance
}

func indexHandler(w http.ResponseWriter, r *http.Request, ctx Context) {
	fmt.Println("index handler")
	fmt.Fprintf(w, "an index")
}

func registerForm(w http.ResponseWriter, req *http.Request) {
	tmpl := getTemplate("register.html")
	tmpl.Execute(w, nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request, ctx Context) {
	if r.Method == "GET" {
		fmt.Println("showing register form")
		registerForm(w, r)
		return
	}
	if r.Method == "POST" {
		username, password, pass2 := r.FormValue("username"), r.FormValue("password"), r.FormValue("pass2")
		if password != pass2 {
			fmt.Fprintf(w, "passwords don't match")
			return
		}
		user, err := ctx.P.CreateUser(username, password)

		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "could not create user")
			return
		}

		sess, _ := store.Get(r, "finch")
		sess.Values["user"] = user.Username
		sess.Save(r, w)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func loginForm(w http.ResponseWriter, req *http.Request) {
	tmpl := getTemplate("login.html")
	tmpl.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request, ctx Context) {
	if r.Method == "GET" {
		loginForm(w, r)
		return
	}
	if r.Method == "POST" {
		username, password := r.FormValue("username"), r.FormValue("password")
		user, found := ctx.P.GetUser(username)

		if !found {
			fmt.Fprintf(w, "user not found")
			return
		}
		if !user.CheckPassword(password) {
			fmt.Fprintf(w, "login failed")
			return
		}

		// store userid in session
		sess, _ := store.Get(r, "finch")
		sess.Values["user"] = user.Username
		sess.Save(r, w)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func logoutHandler(w http.ResponseWriter, r *http.Request, ctx Context) {
	sess, _ := store.Get(r, "finch")
	delete(sess.Values, "user")
	sess.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func getTemplate(filename string) *template.Template {
	var t = template.New("base.html")
	return template.Must(t.ParseFiles(
		filepath.Join(template_dir, "base.html"),
		filepath.Join(template_dir, filename),
	))
}
