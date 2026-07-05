package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev_secret_change_me"
	}

	claims := jwt.MapClaims{
		"sub":  "11111111-1111-1111-1111-111111111111",
		"role": "buyer",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Fatalf("failed to sign token: %v", err)
	}

	fmt.Println(tokenString)
}