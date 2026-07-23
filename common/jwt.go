package common

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Claim struct {
	jwt.StandardClaims
	UserID uint
}

// 这里为什么必须用声明？
var jwt_key = []byte("MyNameIsJohn")

func ReleaseToken(userID int) (string, error) {
	expirationTime := time.Now().Add(time.Hour * 7 * 24)
	claim := Claim{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "127.0.0.1",
			Subject:   "user token",
		},
		UserID: uint(userID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenString, err := token.SignedString(jwt_key)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseToken(tokenString string) (*jwt.Token, *Claim, error) {
	claim := &Claim{}
	token, err := jwt.ParseWithClaims(tokenString, claim, func(t *jwt.Token) (interface{}, error) {
		return jwt_key, nil
	})
	return token, claim, err
}
