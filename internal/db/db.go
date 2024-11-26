package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Storage interface {
	Close()
	GetAllItems(context.Context) ([]TodoItem, error)
	InsertItem(context.Context, TodoItem) (*TodoItem, error)
	DeleteItem(context.Context, int) error
	UpdateItem(context.Context, int, TodoItem) (*TodoItem, error)
}

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

type TodoItem struct {
	Id        int       `db:"id"`
	Task      string    `db:"task"`
	Done      bool      `db:"status"`
	CreatedAt time.Time `db:"created_at"`
}

type DB struct {
	pool *pgxpool.Pool
}

func New(cfg Config) (*DB, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}
	return &DB{
		pool: pool,
	}, nil
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) GetAllItems(ctx context.Context) ([]TodoItem, error) {
	query := `SELECT * FROM todo_items`
	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TodoItem
	for rows.Next() {
		var item TodoItem
		if err := rows.Scan(&item.Id, &item.Task, &item.Done, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (db *DB) InsertItem(ctx context.Context, todo TodoItem) (*TodoItem, error) {
	query := `INSERT INTO todo_items (task, status) VALUES ($1, $2)`
	r, err := db.pool.Exec(ctx, query, todo.Task, todo.Done)
	if err != nil {
		return nil, err
	}
	if r.RowsAffected() == 0 {
		return nil, fmt.Errorf("failed to insert item")
	}
	var createdTodo TodoItem
	selectQuery := `SELECT * FROM todo_items WHERE task = $1`
	err = db.pool.QueryRow(ctx, selectQuery, todo.Task).Scan(&createdTodo.Id, &createdTodo.Task, &createdTodo.Done, &createdTodo.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &createdTodo, nil
}

func (db *DB) DeleteItem(ctx context.Context, id int) error {
	query := `DELETE FROM todo_items WHERE id = $1`
	_, err := db.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) UpdateItem(ctx context.Context, id int, item TodoItem) (*TodoItem, error) {
	query := `UPDATE todo_items SET task=$2, status=$3 WHERE id=$1`
	_, err := db.pool.Exec(ctx, query, id, item.Task, item.Done)
	if err != nil {
		return nil, err
	}
	return &TodoItem{
		Id:   id,
		Task: item.Task,
		Done: item.Done,
	}, nil
}
