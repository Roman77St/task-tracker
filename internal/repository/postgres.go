package repository

import (
	"context"
	"os"
	"task-traker/internal/domain"

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
	return  nil, nil
}
