package service

import (
	"context"
	"fmt"
	"task-traker/internal/domain"
	"time"
)

type TaskService struct {
	Repo domain.TaskRepository
}
func (t TaskService) CreateTask(ctx context.Context, userID int64, title, deadlineStr string) error {
	var task domain.Task
	deadline, err := ParseTime(deadlineStr)
	if err != nil {
		return err
	}
	if time.Since(deadline) > 0 {
		return fmt.Errorf("время выполнения не должно быть в прошлом")
	}
	task.UserID = userID
	task.Title = title
	task.Deadline = deadline

	err = t.Repo.Create(ctx, &task)

	return  err
}

func ParseTime(s string) (time.Time, error) {
	// str := "15.02.2026 11:20"
	const layout = "2.1.2006 15:04"
	parsedTime, err := time.Parse(layout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("некорректно введенная строка времени, %v", err)
	}
	return  parsedTime, nil
}