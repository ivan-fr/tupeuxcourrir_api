package utils

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4/middleware"
)

var secret = "mySecret"

type JwtCustomClaims struct {
	UserID int `json:"id"`
	jwt.StandardClaims
}

func (jCC *JwtCustomClaims) GetToken() string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jCC)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(secret))

	if err != nil {
		panic(err)
	}

	return t
}

var JWTConfig = middleware.JWTConfig{
	Claims:     &JwtCustomClaims{},
	SigningKey: []byte(secret),
	ContextKey: "JWTContext",
}
