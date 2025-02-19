package main

import "net/http"

func addRoutes(
	mux *http.ServeMux,
	templateDir string,
	mediaDir string,
	s *site,
	p *persistence,

) {
	mux.HandleFunc("/", makeHandler(indexHandler, s))
	mux.HandleFunc("/healthz/", makeHandler(healthzHandler, s))
	mux.HandleFunc("GET /post/", makeHandler(postFormHandler, s))
	mux.HandleFunc("POST /post/", makeHandler(postHandler, s))
	mux.HandleFunc("/search/", makeHandler(searchHandler, s))

	mux.HandleFunc("GET /u/{username}/", makeHandler(userIndex, s))
	mux.HandleFunc("GET /u/{username}/feed/", makeHandler(userFeed, s))
	mux.HandleFunc("GET /u/{username}/p/{puuid}/", makeHandler(individualPostHandler, s))
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
