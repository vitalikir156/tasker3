package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/vitalikir156/tasker3/internal/config"
)

var ErrTaskNotFound = errors.New("task with ID not found")
var ErrFreeIDNotFound = errors.New("free ID not found")
var ErrBadStatus = errors.New("Invalid value for status field")


type repository struct {
	pool *pgxpool.Pool
}

type Repository interface {
	CreateTask(ctx context.Context, task Task) (int, error)
	GetTask(ctx context.Context, id int) (Task, error)
	GetTasks(ctx context.Context) ([]Task, error)
	UpdateTask(ctx context.Context, task Task) (error)
	DeleteTask(ctx context.Context, id int) (error)
}

func NewRepository(ctx context.Context, conf config.PostgreSQL) (Repository, error) {
	connString := fmt.Sprintf(
		`user=%s password=%s host=%s port=%d dbname=%s sslmode=%s
        pool_max_conns=%d pool_max_conn_lifetime=%s pool_max_conn_idle_time=%s`,
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.Name,
		conf.SSLMode,
		conf.PoolMaxConns,
		conf.PoolMaxConnLifetime.String(),
		conf.PoolMaxConnIdleTime.String(),
	)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse PostgreSQL config")
	}

	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create PostgreSQL connection pool")
	}

	return &repository{pool}, nil
}


func (r *repository) CreateTask(ctx context.Context, task Task) (int, error) {
if len(task.Status)==0{task.Status = "new"}
	if task.Status != "new" && task.Status != "in_progress" && task.Status != "done" {
		return -1, ErrBadStatus
	}
	query := "INSERT INTO tasks (title, description, status, user_id) VALUES ($1, $2, $3, $4) RETURNING id"
	var id int
	err := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UID).Scan(&id)
	if err != nil {
		return -1, errors.Wrap(err, "failed to insert task")
	}
	return id, nil
}


func (r *repository) GetTask(ctx context.Context, id int) (Task, error) {
	var task Task
	query := "SELECT id, title, description, status, created_at, user_id from tasks where id = $1"
	err := r.pool.QueryRow(ctx, query, id).Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.Created, &task.UID)
	if err != nil {
		return Task{}, errors.Wrap(err, "failed to get task")
	}

	return task, nil
}

func (r *repository) GetTasks(ctx context.Context) ([]Task, error) {
	query := "SELECT id, title, description, status, created_at, user_id from tasks"
	rows, err := r.pool.Query(context.Background(), query)
	if err != nil {
		return []Task{},errors.Wrap(err, "query fault")
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var task Task
		err = rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.Created, &task.UID)
		if err != nil {
			return []Task{},errors.Wrap(err, "query scan fault")
		}
		tasks = append(tasks, task)
	}
	return tasks,nil
}

func (r *repository) UpdateTask(ctx context.Context, task Task) (error) {
	if task.Status != "new" && task.Status != "in_progress" && task.Status != "done" {
		return ErrBadStatus
	}
	query := "UPDATE tasks SET title = $1, description = $2, status = $3, user_id = $4 where id=$5"
	out, err := r.pool.Exec(context.Background(), query,
		task.Title, task.Description, task.Status, task.UID, task.ID)
	if err != nil {
		return errors.Wrap(err, "update fault")
	}
	if out.RowsAffected() < 1 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *repository) DeleteTask(ctx context.Context, id int) (error) {
	query := "DELETE FROM tasks where id=$1"
	out, err := r.pool.Exec(context.Background(), query, id)
	if err != nil {
		return errors.Wrap(err, "delete fault")
	}
	if out.RowsAffected() < 1 {
		return ErrTaskNotFound
	}

	return nil
}