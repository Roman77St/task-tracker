package service

import (
	"context"
	"fmt"
	"task-traker/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const TIME_FORMAT = "02.01.2006 15:04"

func TestParseTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid", "15.02.2026 11:20", false},
		{"Invalid format", "2026-02-15 11:20", true},
		{"Invalid format", "15.02.2026", true},
		{"Empty string", "", true},
		{"Gibberish", "apple-pie", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTime(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type MockRepo struct {
	errToReturn    error
	saveCalled     bool
	savedUserID    int64
	savedCode      string
	saveCodeCalled bool
}

func (m *MockRepo) Create(ctx context.Context, task *domain.Task) error {
	m.saveCalled = true
	return m.errToReturn
}

func (m *MockRepo) GetTasksByUserID(ctx context.Context, userID int64) ([]domain.Task, error) {
	return nil, nil
}
func (m *MockRepo) GetActiveTasks(ctx context.Context) ([]domain.Task, error) { return nil, nil }
func (m *MockRepo) MarkAsNotified(ctx context.Context, taskID int) error      { return nil }
func (m *MockRepo) DeleteByID(ctx context.Context, id string) error           { return nil }
func (m *MockRepo) VerifyAuthCode(ctx context.Context, userID int64, code string) (bool, error) {
	return true, nil
}

func TestCreateTask_RepoError(t *testing.T) {
	mock := &MockRepo{errToReturn: fmt.Errorf("database connection lost")}
	s := TaskService{Repo: mock}

	futureTime := time.Now().Add(time.Hour).Format(TIME_FORMAT)

	err := s.CreateTask(context.Background(), 123, "Тестовая задача", futureTime)

	assert.Error(t, err)
	assert.Equal(t, "database connection lost", err.Error())
}

func TestCreateTask_IncorrectTime(t *testing.T) {

	futureTime := time.Now().Add(10 * time.Hour).Format(TIME_FORMAT)
	pastTime := time.Now().Add(-10 * time.Hour).Format(TIME_FORMAT)

	tests := []struct {
		name      string
		inputTime string
		wantErr   bool
	}{
		{"Valid time", futureTime, false},
		{"time in the past", pastTime, true},
	}
	mock := &MockRepo{}
	s := TaskService{
		Repo: mock,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.CreateTask(context.Background(), 123, "Test task", tt.inputTime)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateTask_Success(t *testing.T) {
	mock := &MockRepo{}

	s := &TaskService{
		Repo: mock,
	}
	futureTime := time.Now().Add(time.Hour).Format(TIME_FORMAT)

	err := s.CreateTask(context.Background(), 123, "Купить хлеб", futureTime)

	assert.NoError(t, err)
	assert.True(t, mock.saveCalled)
}
