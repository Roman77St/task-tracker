package repository

import (
	"context"
	"fmt"
	"os"

	"task-traker/internal/domain"

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
		WHERE notified = false
			AND deadline <= (NOW() + INTERVAL '15 minutes');
	`
	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка GetActiveTasks: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.Task])
}


func (r *Repository) GetTasksByUserID(ctx context.Context, userID int64) ([]domain.Task, error) {
	query := `
	SELECT id, user_id, title, deadline, notified, created_at
	FROM tasks
	WHERE notified = false AND user_id = $1
	ORDER BY deadline;
	`
	rows, err := r.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка GetTasksByUserID: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.Task])
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
