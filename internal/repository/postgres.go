package repository

import (
	"context"
	"fmt"
	"os"
	"task-traker/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	DB *pgxpool.Pool
}

func InitDB(ctx context.Context) (*Repository, error) {
	connStr := os.Getenv("DATABASE_URL")
	conn, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}
	err = conn.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return &Repository{DB: conn}, nil
}

func (r *Repository) Create(ctx context.Context, task *domain.Task) error {
	query := `
	INSERT INTO tasks (user_id, title, deadline)
	VALUES ($1, $2, $3);
	`
	_, err := r.DB.Exec(
		ctx,
		query,
		task.UserID,
		task.Title,
		task.Deadline)

	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetActiveTasks(ctx context.Context) ([]domain.Task, error) {
	query := `
		SELECT id, user_id, title, deadline, notified, created_at
		FROM tasks
		WHERE notified = false AND deadline <= $1;
	`
	rows, err := r.DB.Query(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("ошибка GetActiveTasks: %w", err)
	}
	defer rows.Close()

	activeTasks, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Task])
	if err != nil {
		return nil, fmt.Errorf("failed to collect rows: %w", err)
	}

	return activeTasks, nil
}

func (r *Repository) MarkAsNotified(ctx context.Context, taskID int) error {
	query := `UPDATE tasks SET notified = true
			  WHERE id = $1;`
	_, err := r.DB.Exec(ctx, query, taskID)
	if err != nil {
		return fmt.Errorf("marcAsNotified error: %v", err)
	}
	return nil
}
