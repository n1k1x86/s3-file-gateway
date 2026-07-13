package middleware

import (
	"net/http"

	"github.com/n1k1x86/libs/http_server"

	"go.uber.org/zap"
)

func Logging(logger *zap.Logger) http_server.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info(r.Method + " " + r.Pattern)
			next.ServeHTTP(w, r)
		})
	}
}
