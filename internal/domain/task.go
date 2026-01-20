package domain

import (
	"context"
	"time"
)

type Task struct {
	ID        int       `db:"id"`
	UserID    int64     `db:"user_id"`
	Title     string    `db:"title"`
	Deadline  time.Time `db:"deadline"`
	Notified  bool      `db:"notified"`
	CreatedAt time.Time `db:"created_at"`
}

type TaskRepository interface {
	Create(context.Context, *Task) error
	GetActiveTasks(context.Context) ([]Task, error)
	MarkAsNotified(context.Context, int) error
	GetTasksByUserID(context.Context, int64) ([]Task, error)
	DeleteByID(context.Context, string) error
	SaveAuthCode(context.Context, int64, string, time.Time) error
	VerifyAuthCode(context.Context, int64, string) (bool, error)
}
