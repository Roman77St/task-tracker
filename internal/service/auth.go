package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func (s *TaskService) GenerateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(), // Токен на 3 дня
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtSecret)
}

func (s *TaskService) VerifyToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) { return jwtSecret, nil })

	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}

	userID := int64(claims["user_id"].(float64))
	return userID, nil
}

func (s *TaskService) GenerateAuthCode(ctx context.Context, userID int64) (string, error) {
	// 1. Генерируем 6 случайных цифр
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	expiry := time.Now().Add(5 * time.Minute)

	err := s.Repo.SaveAuthCode(ctx, userID, code, expiry)

	return code, err
}
