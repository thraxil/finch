package main

import "net/http"

func makeHandler(fn func(http.ResponseWriter, *http.Request, *site), s *site) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, s)
	}
}

func addRoutes(
	mux *http.ServeMux,
	templateDir string,
	mediaDir string,
	s *site,
	p *persistence,

) {
	mux.Handle("/", indexHandler(s))
	mux.HandleFunc("/healthz/", healthzHandler)
	mux.Handle("GET /post/", postFormHandler(s))
	mux.Handle("POST /post/", postHandler(s))
	mux.Handle("/search/", searchHandler(s))

	mux.Handle("GET /u/{username}/", userIndex(s))
	mux.Handle("GET /u/{username}/feed/", userFeed(s))
	mux.Handle("GET /u/{username}/p/{puuid}/", individualPostHandler(s))
	mux.HandleFunc("POST /u/{username}/p/{puuid}/delete/", makeHandler(postDelete, s))
	mux.HandleFunc("GET /u/{username}/c/{slug}/", makeHandler(channelIndex, s))
	mux.HandleFunc("GET /u/{username}/c/{slug}/feed/", makeHandler(channelFeed, s))
	mux.HandleFunc("POST /u/{username}/c/{slug}/delete/", makeHandler(channelDelete, s))

	// authy stuff
	mux.HandleFunc("GET /register/", makeHandler(registerFormHandler, s))
	mux.HandleFunc("POST /register/", makeHandler(registerHandler, s))
	mux.HandleFunc("GET /login/", makeHandler(loginFormHandler, s))
	mux.HandleFunc("POST /login/", makeHandler(loginHandler, s))
	mux.HandleFunc("/logout/", makeHandler(logoutHandler, s))

	// static misc.
	mux.HandleFunc("/favicon.ico", faviconHandler)
	mux.Handle("/media/", http.StripPrefix("/media/",
		http.FileServer(http.Dir(mediaDir))))

}
