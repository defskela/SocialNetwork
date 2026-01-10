package v1

import (
	"context"
	"net/http"
	"strings"
)

type CtxKey string

const (
	CtxKeyUserID CtxKey = "userID"
)

func (h *Handler) userIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "empty auth header", http.StatusUnauthorized)
			return
		}

		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			http.Error(w, "invalid auth header", http.StatusUnauthorized)
			return
		}

		if headerParts[1] == "" {
			http.Error(w, "token is empty", http.StatusUnauthorized)
			return
		}

		userID, err := h.services.Auth.ParseToken(headerParts[1])
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), CtxKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
