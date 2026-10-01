package jwtservice

import (
	"crypto/rand"
	"errors"
	"tictactoe/internal/store"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Jwtservice struct {
	signKey    []byte
	repository *store.Repository
	parser     *jwt.Parser
}

func New(r *store.Repository) *Jwtservice {
	key := []byte(rand.Text())
	return &Jwtservice{
		signKey:    key,
		repository: r,
		parser:     jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name})),
	}
}

type TokenType string

const (
	Access  TokenType = "access"
	Refresh TokenType = "refresh"
)

var ErrTokenExpired error = errors.New("error token expired")

type Claims struct {
	jwt.RegisteredClaims
	Type TokenType
}

func (j *Jwtservice) CreateToken(userId string, tokenType TokenType) (token string, err error) {
	var ttl time.Duration
	switch tokenType {
	case Access:
		ttl = time.Minute * 5
	case Refresh:
		ttl = time.Hour * 120
	}

	claims := Claims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userId,
		},
		tokenType,
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	token, err = t.SignedString(j.signKey)
	if err != nil {
		return "", err
	}

	if tokenType == Refresh {
		j.repository.UpdateRefreshToken(userId, token)
	}

	return token, nil
}

func (j *Jwtservice) ValidateToken(tokenString string) (userId string, tokenType TokenType, err error) {
	token, err := j.parser.ParseWithClaims(tokenString,
		&Claims{},
		func(tok *jwt.Token) (any, error) {
			return j.signKey, nil
		})
	if err != nil {
		return "", "", err
	}

	expAt := token.Claims.(*Claims).ExpiresAt.Time
	if expAt.Before(time.Now()) {
		return "", "", ErrTokenExpired
	}

	return token.Claims.(*Claims).Subject, token.Claims.(*Claims).Type, nil
}
