package httpHandler

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		slog.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start))
	})
}

func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		bad, _ := h.service.Redis.GetToken(r.Context(), "blacklist:"+tokenString)
		if bad != "" {
			http.Error(w, "Token is invalidated (logged out)", http.StatusUnauthorized)
			return
		}

		userID, err := h.service.VerifyToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
