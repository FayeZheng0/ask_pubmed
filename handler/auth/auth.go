package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/FayeZheng0/ask_pubmed/handler/render"
	"github.com/FayeZheng0/ask_pubmed/session"
)

func HandleAuthentication(s *session.Session) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			accessToken := getBearerToken(r)
			if accessToken == "" {
				next.ServeHTTP(w, r)
				return
			}

			user, err := s.Login(ctx, getBearerToken(r), false)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r.WithContext(
				session.WithUser(ctx, user),
			))
		}

		return http.HandlerFunc(fn)
	}
}

func Me() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, found := session.UserFrom(ctx)
		if !found {
			render.Error(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
			return
		}
		render.JSON(w, user)
	}
}

func getBearerToken(r *http.Request) string {
	s := r.Header.Get("Authorization")
	return strings.TrimPrefix(s, "Bearer ")
}
