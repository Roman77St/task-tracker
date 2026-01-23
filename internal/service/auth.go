package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *TaskService) GenerateToken(userID int64, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(duration).Unix(),
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
	// Генерируем 6 случайных цифр
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	key := fmt.Sprintf("otp:%d", userID)

	err := s.Redis.SetToken(ctx, key, code, time.Minute*5)

	return code, err
}

func (s *TaskService) Login(ctx context.Context, userID int64, code string) (*TokenPair, error) {
	key := fmt.Sprintf("otp:%d", userID)
	savedCode, err := s.Redis.GetToken(ctx, key)
	if err != nil || savedCode != code {
		return nil, errors.New("invalid or expired code")
	}

	accessToken, err := s.GenerateToken(userID, time.Minute*15)
	if err != nil {
		return nil, err
	}

	refreshToken := uuid.New().String()
	refreshKey := fmt.Sprintf("refresh:%s", refreshToken)
	err = s.Redis.SetToken(ctx, refreshKey, userID, time.Hour*24*30)
	if err != nil {
		return nil, err
	}

	s.Redis.DeleteToken(ctx, key)

	return &TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *TaskService) Logout(ctx context.Context, accessToken string, refreshToken string) error {
	// Удаляем Refresh токен из Redis (он больше не валиден)
	refreshKey := fmt.Sprintf("refresh:%s", refreshToken)
	s.Redis.DeleteToken(ctx, refreshKey)

	// TTL должен быть равен остатку времени жизни токена
	blacklistKey := fmt.Sprintf("blacklist:%s", accessToken)
	return s.Redis.SetToken(ctx, blacklistKey, "true", time.Minute*15)
}

func (s *TaskService) Refresh(ctx context.Context, oldRefreshToken string) (*TokenPair, error) {
	refreshKey := fmt.Sprintf("refresh:%s", oldRefreshToken)

	// 1. Ищем Refresh токен в Redis
	val, err := s.Redis.GetToken(ctx, refreshKey)
	if err != nil {
		return nil, errors.New("refresh token expired or invalid")
	}

	// 2. Достаем userID из значения
	var userID int64
	fmt.Sscanf(val, "%d", &userID)

	// 3. Генерируем новую пару
	accessToken, _ := s.GenerateToken(userID, time.Minute*15)
	newRefreshToken := uuid.New().String()

	// 4. Удаляем старый и сохраняем новый
	s.Redis.DeleteToken(ctx, refreshKey)
	s.Redis.SetToken(ctx, fmt.Sprintf("refresh:%s", newRefreshToken), userID, time.Hour*24*30)

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
