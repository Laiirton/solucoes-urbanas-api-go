package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				respondJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "Token de acesso não encontrado"})
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				respondJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "Sessão expirada. Faça login novamente."})
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				respondJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "Erro ao validar token"})
				return
			}

			userID, ok := claims["user_id"].(float64)
			if !ok {
				respondJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "Token inválido"})
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, int64(userID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns middleware that restricts access to users with one of the specified roles.
// It fetches the user from DB to check the type. The user must be authenticated first (via Auth middleware).
func RequireRole(userRepo *repository.UserRepository, roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDKey).(int64)
			if !ok {
				respondJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "Não autenticado"})
				return
			}

			user, err := userRepo.GetUserByID(r.Context(), userID)
			if err != nil || user.Type == nil {
				respondJSON(w, http.StatusForbidden, models.ErrorResponse{Error: "Acesso negado"})
				return
			}

			for _, role := range roles {
				if *user.Type == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			respondJSON(w, http.StatusForbidden, models.ErrorResponse{Error: "Você não tem permissão para acessar este recurso"})
		})
	}
}

func respondJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}
