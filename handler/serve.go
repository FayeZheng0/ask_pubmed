package handler

import (
	"fmt"
	"net/http"

	"github.com/FayeZheng0/ask_pubmed/core"
	"github.com/FayeZheng0/ask_pubmed/handler/auth"
	"github.com/FayeZheng0/ask_pubmed/handler/echo"
	"github.com/FayeZheng0/ask_pubmed/handler/render"
	"github.com/FayeZheng0/ask_pubmed/handler/search"
	"github.com/FayeZheng0/ask_pubmed/session"

	"github.com/go-chi/chi"
)

func New(cfg Config,
	session *session.Session,

	searchz core.SearchResultService,
) Server {
	return Server{
		cfg:     cfg,
		session: session,

		searchz: searchz,
	}
}

type (
	Config struct {
	}

	Server struct {
		cfg     Config
		session *session.Session

		searchz core.SearchResultService
	}
)

func (s Server) HandleRest() http.Handler {
	r := chi.NewRouter()
	r.Use(render.WrapResponse(true))
	r.Use(auth.HandleAuthentication(s.session))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Error(w, http.StatusNotFound, fmt.Errorf("not found"))
	})

	r.Route("/echo", func(r chi.Router) {
		r.Get("/{msg}", echo.Get())
		r.Post("/", echo.Post())
	})

	r.With(s.LoginRequired()).Route("/me", func(r chi.Router) {
		r.Get("/", auth.Me())
	})

	r.Route("/search", func(r chi.Router) {
		r.Get("/{query}", search.Get(s.searchz))
		r.Post("/", search.Post(s.searchz))
	})

	return r
}

func (s Server) LoginRequired() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if _, found := session.UserFrom(ctx); !found {
				render.Error(w, http.StatusUnauthorized, fmt.Errorf("session unauthorized"))
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
