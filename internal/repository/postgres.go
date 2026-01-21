package repository

import (
	"context"
	"fmt"
	"os"
	"time"

	"task-traker/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	DB *pgxpool.Pool
}

func InitDB(ctx context.Context) (*Repository, error) {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST") // В Docker Compose это будет "db"
	dbPort := os.Getenv("DB_PORT") // В Docker Compose это будет "5432"
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPass, dbHost, dbPort, dbName)

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

// func (r *Repository) Delete(ctx context.Context, userID int64, taskNumber int) error {
// 	offset := taskNumber - 1
// 	if offset < 0 {
// 		return fmt.Errorf("неверный номер задачи")
// 	}

// 	query := `
// 	DELETE FROM tasks
// 	WHERE id = (
// 	SELECT id
// 	FROM tasks
// 	WHERE notified = false AND user_id = $1
// 	ORDER BY deadline
// 	LIMIT 1
// 	OFFSET $2);
// 	`
// 	res, err := r.DB.Exec(ctx, query, userID, taskNumber)
// 	if err != nil {
// 		return fmt.Errorf("Delete: %w", err)
// 	}
// 	if res.RowsAffected() == 0 {
// 		return fmt.Errorf("задача под номером %d не найдена", taskNumber)
// 	}
// 	return nil
// }

func (r *Repository) DeleteByID(ctx context.Context, id string) error {
	query := "DELETE FROM tasks WHERE id = $1;"
	_, err := r.DB.Exec(ctx, query, id)
	return err
}

func (r *Repository) SaveAuthCode(ctx context.Context, userID int64, code string, expiry time.Time) error {
	query := `
        INSERT INTO auth_codes (user_id, code, expires_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id) DO UPDATE
        SET code = EXCLUDED.code, expires_at = EXCLUDED.expires_at
    `
	_, err := r.DB.Exec(ctx, query, userID, code, expiry)
	return err
}

func (r *Repository) VerifyAuthCode(ctx context.Context, userID int64, code string) (bool, error) {
	var dbCode string
	var expiresAt time.Time

	query := `SELECT code, expires_at FROM auth_codes WHERE user_id = $1`
	err := r.DB.QueryRow(ctx, query, userID).Scan(&dbCode, &expiresAt)

	if err != nil {
		return false, err
	}

	if dbCode != code || time.Now().After(expiresAt) {
		return false, nil
	}

	r.DB.Exec(ctx, "DELETE FROM auth_codes WHERE user_id = $1", userID)

	return true, nil
}
