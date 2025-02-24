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
	"sync"
	"time"

	"github.com/braintree/manners"
	"github.com/gorilla/sessions"
)

var templateDir = "templates"

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

func run(ctx context.Context,
	w io.Writer,
	args []string,
	getenv func(string) string,
	stdin io.Reader,
	stdout, stderr io.Writer,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	log.Println("Starting Finch...")
	// set up the database file
	p := newPersistence(getenv("FINCH_DB_FILE"))
	defer p.Close()
	templateDir = getenv("FINCH_TEMPLATE_DIR")
	mediaDir := getenv("FINCH_MEDIA_DIR")
	s := newSite(
		p,
		getenv("FINCH_BASE_URL"),
		sessions.NewCookieStore([]byte(getenv("FINCH_SECRET"))),
		getenv("FINCH_ITEMS_PER_PAGE"),
		getenv("FINCH_ALLOW_REGISTRATION"),
	)
	srv := NewServer(
		templateDir,
		mediaDir,
		s,
		p,
	)
	httpServer := manners.NewServer()
	httpServer.Addr = net.JoinHostPort("", getenv("FINCH_PORT"))
	httpServer.Handler = srv

	go func() {
		log.Println("running on " + getenv("FINCH_PORT"))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(stderr, "error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
	return nil

}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args, os.Getenv, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
