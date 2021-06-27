package auth

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"kontrakt-server/prisma/db"
	"kontrakt-server/utils"
	"net/http"
	"strings"
)

var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

func Middleware(prisma *db.PrismaClient) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")

			// Allow unauthenticated users in
			if len(tokenString) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
			claims, err := utils.VerifyToken(tokenString)
			if err != nil {
				http.Error(w, "Error verifying JWT token: "+err.Error(), http.StatusUnauthorized)
				return
			}

			username, ok := claims.(jwt.MapClaims)["username"].(string)
			if !ok {
				http.Error(w, "Invalid user", http.StatusForbidden)
				return
			}
			user, err := prisma.User.FindUnique(db.User.Username.Equals(username)).Exec(r.Context())
			if err != nil {
				http.Error(w, "Invalid user", http.StatusForbidden)
				return
			}

			//// put it in context
			ctx := context.WithValue(r.Context(), userCtxKey, user)

			// and call the next with our new context
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func ForContext(ctx context.Context) *db.UserModel {
	raw, _ := ctx.Value(userCtxKey).(*db.UserModel)
	return raw
}
