package main // import "github.com/thraxil/finch"

import (
	"context"
	_ "expvar"
	"fmt"
	"io"
	"log"
	"net"
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

func NewServer(
	templateDir string,
	mediaDir string,
	s *site,
	p *persistence,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(
		mux,
		templateDir,
		mediaDir,
		s,
		p,
	)
	var handler http.Handler = mux
	handler = LoggingHandler(mux)
	return handler
}

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	log.Println("Starting Finch...")
	// set up the database file
	p := newPersistence(os.Getenv("FINCH_DB_FILE"))
	defer p.Close()
	templateDir = os.Getenv("FINCH_TEMPLATE_DIR")
	mediaDir := os.Getenv("FINCH_MEDIA_DIR")
	s := newSite(
		p,
		os.Getenv("FINCH_BASE_URL"),
		sessions.NewCookieStore([]byte(os.Getenv("FINCH_SECRET"))),
		os.Getenv("FINCH_ITEMS_PER_PAGE"),
		os.Getenv("FINCH_ALLOW_REGISTRATION"),
	)
	srv := NewServer(
		templateDir,
		mediaDir,
		s,
		p,
	)
	httpServer := manners.NewServer()
	httpServer.Addr = net.JoinHostPort("", os.Getenv("FINCH_PORT"))
	httpServer.Handler = srv

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
