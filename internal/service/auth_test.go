package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJWTFlow(t *testing.T) {

	s := &TaskService{}
	userID := int64(123)

	token, err := s.GenerateToken(userID)
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

func TestGenerateAuthCode_Success(t *testing.T) {
	mock := &MockRepo{}
	s := &TaskService{Repo: mock}

	userID := int64(1188444466)

	code, err := s.GenerateAuthCode(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, code, 6)
	assert.True(t, mock.saveCodeCalled, "Метод SaveAuthCode должен быть вызван")
	assert.Equal(t, userID, mock.savedUserID)
	assert.Equal(t, code, mock.savedCode)
}

func TestVerifyToken_Invalid(t *testing.T) {
	s := &TaskService{}

	_, err := s.VerifyToken("invalid.token.string")
	assert.Error(t, err)
	assert.Equal(t, "invalid token", err.Error())
}
