package auth

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
)

type UserClaims struct {
	ID     string      `json:"id"`
	Type   string      `json:"type"` // user, service
	Role   string      `json:"role"`
	Scope  string      `json:"scope"`
	Object interface{} `json:"object"` // user, service, loadbalancer, server
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authDisabled := os.Getenv("AUTH_ENABLED") == "false" || os.Getenv("AUTH_ENABLED") == "0"
		if authDisabled {
			next.ServeHTTP(w, r)
			return
		}

		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")

		// Read the public key from the file
		publicKey, err := os.ReadFile(os.Getenv("SECNEX_GATEWAY_PUBLIC_KEY"))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Parse the token
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return publicKey, nil
		})
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !parsedToken.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userClaims := UserClaims{
			ID:     claims["id"].(string),
			Type:   claims["type"].(string),
			Role:   claims["role"].(string),
			Scope:  claims["scope"].(string),
			Object: claims["object"],
		}

		ctx := context.WithValue(r.Context(), "userClaims", userClaims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
