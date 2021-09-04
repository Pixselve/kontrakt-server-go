package utils

import (
	jwt "github.com/dgrijalva/jwt-go"
	"os"
)

func GetToken(username string) (string, error) {
	value := os.Getenv("JWT_KEY")
	signingKey := []byte(value)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
	})
	tokenString, err := token.SignedString(signingKey)
	return tokenString, err
}

func VerifyToken(tokenString string) (jwt.Claims, error) {
	value := os.Getenv("JWT_KEY")
	signingKey := []byte(value)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		return nil, err
	}
	return token.Claims, err
}
