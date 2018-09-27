package internal

import (
	"fmt"
	"log"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	request "github.com/dgrijalva/jwt-go/request"
)

func ValidateJWT(r *http.Request) (bool, string) {

	token, err := request.ParseFromRequest(r, request.OAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		} else {
			return []byte("my-secret-token-to-change-in-production"), nil
		}
	})

	if err == nil && token.Valid {
		claims := token.Claims.(jwt.MapClaims)
		return true, claims["sub"].(string)
	} else {
		log.Println(err)
		return false, ""
	}
}
