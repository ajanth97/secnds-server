package token

import (
	"log"
	"secnds-server/env"
	"time"

	"github.com/golang-jwt/jwt"
)

const jwt_secret = "JWT_SECRET"

var jwtSecret []byte = env.GetByte(jwt_secret)

type jwtCustomClaims struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	jwt.StandardClaims
}

func GetJwtToken(userId string, userEmail string) string {
	claims := &jwtCustomClaims{
		userId,
		userEmail,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Fatalf("Couldn't sign token %v", err)
	}
	return t
}
