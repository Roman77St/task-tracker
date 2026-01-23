package service

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"task-traker/internal/repository"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestJWTFlow(t *testing.T) {

	s := &TaskService{}
	userID := int64(123)

	token, err := s.GenerateToken(userID, time.Minute*5)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedID, err := s.VerifyToken(token)
	assert.NoError(t, err)
	assert.Equal(t, parsedID, userID)
}

func (m *MockRepo) SaveAuthCode(ctx context.Context, userID int64, code string, expiry time.Time) error {
	m.saveCodeCalled = true
	m.savedUserID = userID
	m.savedCode = code
	return m.errToReturn
}

func TestVerifyToken_Invalid(t *testing.T) {
	s := &TaskService{}

	_, err := s.VerifyToken("invalid.token.string")
	assert.Error(t, err)
	assert.Equal(t, "invalid token", err.Error())
}

func TestGenerateAuthCode_Success(t *testing.T) {
	// Запускаем виртуальный Redis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	// Создаем репозиторий Redis, указывая адрес виртуального сервера
	redisRepo := repository.NewRedisRepo(mr.Addr())

	// Создаем сервис и передаем ему redisRepo
	s := &TaskService{
		Redis: redisRepo,
		// Repo: &MockRepo{}, // если метод использует и Postgres, добавить мок
	}
	userID, _ := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
	ctx := context.Background()

	code, err := s.GenerateAuthCode(ctx, userID)

	// Проверки
	assert.NoError(t, err)
	assert.Len(t, code, 6)

	// Проверяем, что в miniredis действительно создался ключ
	expectedKey := fmt.Sprintf("otp:%d", userID)
	assert.True(t, mr.Exists(expectedKey), "Ключ должен существовать в Redis")

	val, _ := mr.Get(expectedKey)
	assert.Equal(t, code, val, "Код в Redis должен совпадать с сгенерированным")
}