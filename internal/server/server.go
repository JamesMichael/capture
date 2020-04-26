package server

import (
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/jamesmichael/capture/internal/config"
	"github.com/justinas/nosurf"
	"go.uber.org/zap"
)

type Server interface {
	Serve() error
}

type server struct {
	logger   *zap.SugaredLogger
	addr     string
	template *template.Template
	server   *http.Server
	config   *config.Config
}

func New(options ...Option) (Server, error) {
	s := &server{}

	for _, opt := range options {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *server) Serve() error {
	s.logger.Infow("starting server",
		"port", s.addr,
	)

	return s.httpServer().ListenAndServe()
}

func (s *server) httpServer() *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/capture", s.capture())
	mux.HandleFunc("/", s.index())

	return &http.Server{
		Addr:    s.addr,
		Handler: nosurf.New(mux),
	}
}

func (s *server) index() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			s.logger.Infow("invalid request method",
				"method", r.Method,
				"url", r.URL,
			)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != "/" {
			s.logger.Infow("page not found",
				"path", r.URL.Path,
				"url", r.URL,
			)
			http.NotFound(w, r)
			return
		}

		s.template.ExecuteTemplate(w, "index.html", map[string]interface{}{
			"contexts": s.config.ListContexts(),
			"token":    nosurf.Token(r),
		})
	}
}

func (s *server) capture() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != "/capture" {
			http.NotFound(w, r)
			return
		}

		context := r.PostFormValue("context")
		message := r.PostFormValue("message")

		filename := s.config.FileForContext(context)
		if filename == "" {
			s.logger.Infow("invalid context")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		fmt.Fprintf(f, "\n----\n\n%s\n", message)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

type Option func(s *server) error

func WithAddress(addr string) Option {
	return func(s *server) error {
		s.addr = addr
		return nil
	}
}

func WithConfig(path string) Option {
	return func(s *server) error {
		c, err := config.NewFromFile(path)
		if err != nil {
			return err
		}
		s.config = c
		return nil
	}
}

func WithLogger(logger *zap.SugaredLogger) Option {
	return func(s *server) error {
		s.logger = logger
		return nil
	}
}

func WithTemplate(path string) Option {
	return func(s *server) error {
		t, err := template.ParseFiles(path)
		if err != nil {
			return fmt.Errorf("unable to load template file '%s': %w", path, err)
		}
		s.template = t
		return nil
	}
}
