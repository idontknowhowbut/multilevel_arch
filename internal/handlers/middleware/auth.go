package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"tictactoe/internal/app"
	jwtservice "tictactoe/internal/jwt"
)

type AuthService struct {
	service *app.Service
}

func New(service *app.Service) *AuthService {
	return &AuthService{
		service: service,
	}
}

func (a *AuthService) Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		fmt.Println("Auth: token:")
		fmt.Println(token[1])

		userId, tokenType, err := a.service.Jwtservice.ValidateToken(token[1])

		if err != nil {
			fmt.Print(err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if tokenType != jwtservice.Access {
			fmt.Print(err)
			http.Error(w, "Incorrect token type", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(context.Background(), "userId", userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
