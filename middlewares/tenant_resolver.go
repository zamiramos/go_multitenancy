package middlewares

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"net/http"
)

const RealmIdContext string = "realmId"

func CreateTenantResolverMiddleware(logger log.Logger) mux.MiddlewareFunc {
	return mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			realmId := r.Header.Get("Realm-Id")
			if realmId == "" {
				realmId = "default"
			}
			ctx := r.Context()
			ctx = context.WithValue(ctx, RealmIdContext, realmId)

			logger.Log(RealmIdContext, realmId)

			// Call the next handler, which can be another middleware in the chain, or the final handler.
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
}
