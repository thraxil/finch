package main // import "github.com/thraxil/finch"

import (
	_ "expvar"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/braintree/manners"
	"github.com/gorilla/sessions"
)

var templateDir = "templates"

func makeHandler(fn func(http.ResponseWriter, *http.Request, *site), s *site) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, s)
	}
}

func LoggingHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		format := "%s - - [%s] \"%s %s %s\" %s\n"
		fmt.Printf(format, r.RemoteAddr, time.Now().Format(time.RFC1123),
			r.Method, r.URL.Path, r.Proto, r.UserAgent())
		h.ServeHTTP(w, r)
	})
}

func main() {
	log.Println("Starting Finch...")
	p := newPersistence(os.Getenv("FINCH_DB_FILE"))
	defer p.Close()
	templateDir = os.Getenv("FINCH_TEMPLATE_DIR")
	s := newSite(
		p,
		os.Getenv("FINCH_BASE_URL"),
		sessions.NewCookieStore([]byte(os.Getenv("FINCH_SECRET"))),
		os.Getenv("FINCH_ITEMS_PER_PAGE"))

	mux := http.NewServeMux()
	mux.HandleFunc("/", makeHandler(indexHandler, s))
	mux.HandleFunc("/post/", makeHandler(postHandler, s))
	mux.HandleFunc("/search/", makeHandler(searchHandler, s))

	mux.HandleFunc("/u/", makeHandler(userDispatch, s))

	// authy stuff
	mux.HandleFunc("/register/", makeHandler(registerHandler, s))
	mux.HandleFunc("/login/", makeHandler(loginHandler, s))
	mux.HandleFunc("/logout/", makeHandler(logoutHandler, s))

	// static misc.
	mux.HandleFunc("/favicon.ico", faviconHandler)
	mux.Handle("/media/", http.StripPrefix("/media/",
		http.FileServer(http.Dir(os.Getenv("FINCH_MEDIA_DIR")))))

	httpServer := manners.NewServer()
	httpServer.Addr = ":" + os.Getenv("FINCH_PORT")
	httpServer.Handler = LoggingHandler(mux)

	errChan := make(chan error, 10)
	go func() {
		log.Println("running on " + os.Getenv("FINCH_PORT"))
		errChan <- httpServer.ListenAndServe()
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-errChan:
			if err != nil {
				log.Fatal(err)
			}
		case s := <-signalChan:
			log.Println(fmt.Sprintf("Captured %v. Exiting...", s))
			httpServer.BlockingClose()
			os.Exit(0)
		}
	}
}
