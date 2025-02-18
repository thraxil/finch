package main // import "github.com/thraxil/finch"

import (
	"context"
	_ "expvar"
	"fmt"
	"io"
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

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	log.Println("Starting Finch...")
	// set up the database file
	p := newPersistence(os.Getenv("FINCH_DB_FILE"))
	defer p.Close()
	templateDir = os.Getenv("FINCH_TEMPLATE_DIR")
	s := newSite(
		p,
		os.Getenv("FINCH_BASE_URL"),
		sessions.NewCookieStore([]byte(os.Getenv("FINCH_SECRET"))),
		os.Getenv("FINCH_ITEMS_PER_PAGE"),
		os.Getenv("FINCH_ALLOW_REGISTRATION"),
	)

	mux := http.NewServeMux()
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

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
