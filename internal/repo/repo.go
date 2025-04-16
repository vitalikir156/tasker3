package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/vitalikir156/tasker3/internal/config"
)

var (
	ErrTaskNotFound   = errors.New("object with ID not found")
	ErrFreeIDNotFound = errors.New("free ID not found")
	ErrBadStatus      = errors.New("Invalid value for status field")
	ErrEmptyName      = errors.New("name field is empty")
)

const taskstatusNew = "new"

type repository struct {
	pool *pgxpool.Pool
}

type Repository interface {
	CreateTask(ctx context.Context, task Task) (int, error)
	GetTask(ctx context.Context, id int) (Task, error)
	GetTasks(ctx context.Context) ([]Task, error)
	GetTasksOverUID(ctx context.Context, uid int) ([]Task, error)
	UpdateTask(ctx context.Context, task Task) error
	DeleteTask(ctx context.Context, id int) error

	CreateUser(ctx context.Context, user User) (int, error)
	GetUser(ctx context.Context, id int) (User, error)
	GetUsers(ctx context.Context) ([]User, error)
	UpdateUser(ctx context.Context, user User) error
	DeleteUser(ctx context.Context, id int) error
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
	if len(task.Status) == 0 {
		task.Status = taskstatusNew
	}
	if task.Status != taskstatusNew && task.Status != "in_progress" && task.Status != "done" {
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
	err := r.pool.QueryRow(ctx, query, id).Scan(&task.ID, &task.Title,
		&task.Description, &task.Status, &task.Created, &task.UID)
	if err != nil {
		return Task{}, errors.Wrap(err, "failed to get task")
	}

	return task, nil
}

func (r *repository) GetTasks(ctx context.Context) ([]Task, error) {
	query := "SELECT id, title, description, status, created_at, user_id from tasks"
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return []Task{}, errors.Wrap(err, "query fault")
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var task Task
		err = rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.Created, &task.UID)
		if err != nil {
			return []Task{}, errors.Wrap(err, "query scan fault")
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *repository) GetTasksOverUID(ctx context.Context, uid int) ([]Task, error) {
	query := "SELECT id, title, description, status, created_at, user_id from tasks where user_id=$1"
	rows, err := r.pool.Query(ctx, query, uid)
	if err != nil {
		return []Task{}, errors.Wrap(err, "query fault")
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var task Task
		err = rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.Created, &task.UID)
		if err != nil {
			return []Task{}, errors.Wrap(err, "query scan fault")
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *repository) UpdateTask(ctx context.Context, task Task) error {
	if task.Status != taskstatusNew && task.Status != "in_progress" && task.Status != "done" {
		return ErrBadStatus
	}
	query := "UPDATE tasks SET title = $1, description = $2, status = $3, user_id = $4 where id=$5"
	out, err := r.pool.Exec(ctx, query,
		task.Title, task.Description, task.Status, task.UID, task.ID)
	if err != nil {
		return errors.Wrap(err, "update fault")
	}
	if out.RowsAffected() < 1 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *repository) DeleteTask(ctx context.Context, id int) error {
	query := "DELETE FROM tasks where id=$1"
	out, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "delete fault")
	}
	if out.RowsAffected() < 1 {
		return ErrTaskNotFound
	}

	return nil
}

func (r *repository) CreateUser(ctx context.Context, user User) (int, error) {
	if len(user.Name) == 0 {
		return -1, ErrEmptyName
	}

	query := "INSERT INTO users (uname) VALUES ($1) RETURNING id"
	var id int
	err := r.pool.QueryRow(ctx, query, user.Name).Scan(&id)
	if err != nil {
		return -1, errors.Wrap(err, "failed to insert user")
	}
	return id, nil
}

func (r *repository) GetUser(ctx context.Context, id int) (User, error) {
	var user User
	query := "SELECT id, uname, taskread, taskwrite, userread, userwrite from users where id = $1"
	err := r.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Name, &user.Taskread,
		&user.Taskwrite, &user.Userread, &user.Userwrite)
	if err != nil {
		return User{}, errors.Wrap(err, "failed to get user")
	}

	return user, nil
}

func (r *repository) GetUsers(ctx context.Context) ([]User, error) {
	query := "SELECT id, uname, taskread, taskwrite, userread, userwrite from users"
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return []User{}, errors.Wrap(err, "query fault")
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var user User
		err = rows.Scan(&user.ID, &user.Name, &user.Taskread, &user.Taskwrite, &user.Userread, &user.Userwrite)
		if err != nil {
			return []User{}, errors.Wrap(err, "query scan fault")
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *repository) UpdateUser(ctx context.Context, user User) error {
	query := "UPDATE users SET uname = $1, taskread = $2, taskwrite = $3, userread = $4, userwrite = $5 where id=$6"
	out, err := r.pool.Exec(ctx, query,
		user.Name, user.Taskread, user.Taskwrite, user.Userread, user.Userwrite, user.ID)
	if err != nil {
		return errors.Wrap(err, "update fault")
	}
	if out.RowsAffected() < 1 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *repository) DeleteUser(ctx context.Context, id int) error {
	query := "DELETE FROM users where id=$1"
	out, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "delete fault")
	}
	if out.RowsAffected() < 1 {
		return ErrTaskNotFound
	}

	return nil
}
