package authorization

import (
	"context"
	"crypto/ecdsa"
	"net/http"

	"github.com/hesoyamTM/apphelper-sso/internal/lib/jwt"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
)

type ctxKey string
type Middleware func(next http.Handler) http.Handler

const (
	Uid ctxKey = "uid"
)

func NewAuthMiddleware(authMethods map[string]bool, publicKey *ecdsa.PublicKey) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := logger.GetLoggerFromCtx(r.Context())

			if !authMethods[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			cookieToken := r.CookiesNamed("authorization")

			if len(cookieToken) == 0 {
				l.Error(r.Context(), "cookies is empty")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			bearerToken := cookieToken[0].Value
			uid, err := jwt.VerifyBearerToken(bearerToken, publicKey)
			if err != nil {
				l.Error(r.Context(), err.Error())
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), Uid, uid))

			next.ServeHTTP(w, r)
		})
	}
}
