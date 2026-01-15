package service

import (
	"context"
	"fmt"
	"log/slog"
	"task-traker/pkg/telegram"
	"time"
)

func (s *TaskService) StartNotificationWorker(ctx context.Context, bot *telegram.Client)  {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <- ctx.Done():
			return
		case <-ticker.C:
			tasks, err := s.Repo.GetActiveTasks(ctx)
			if err != nil {
				slog.Error("worker error", "err", err)
				continue
			}

			for _, task := range tasks {
				msg := fmt.Sprintf("⏰Напоминание: %s", task.Title)
				if err := bot.SendMessage(task.UserID, msg); err == nil {
					s.Repo.MarkAsNotified(ctx, task.ID)
				}
			}
		}
    }
}