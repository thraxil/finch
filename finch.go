package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var store sessions.Store
var template_dir = "templates"

func makeHandler(fn func(http.ResponseWriter, *http.Request, Context), ctx Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, ctx)
	}
}

func main() {
	p := NewPersistance(os.Getenv("FINCH_DB_FILE"))
	defer p.Close()

	store = sessions.NewCookieStore([]byte(os.Getenv("FINCH_SECRET")))

	ctx := Context{P: p}

	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/", makeHandler(indexHandler, ctx))

	http.HandleFunc("/register/", makeHandler(registerHandler, ctx))
	http.HandleFunc("/login/", makeHandler(loginHandler, ctx))
	http.HandleFunc("/logout/", makeHandler(logoutHandler, ctx))
	http.Handle("/media/", http.StripPrefix("/media/",
		http.FileServer(http.Dir(os.Getenv("FINCH_MEDIA_DIR")))))
	log.Fatal(http.ListenAndServe(":"+os.Getenv("FINCH_PORT"), nil))
}
