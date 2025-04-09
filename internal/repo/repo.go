package repo

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

// Слой репозитория, здесь должны быть все методы, связанные с базой данных

// SQL-запрос на вставку задачи
const (
	insertTaskQuery = `INSERT INTO tasks (title, description) VALUES ($1, $2) RETURNING id;`
	getTaskQuery    = `select id, title, description, status, created_at, updated_at from tasks where id=$1;`
)

type repository struct {
	mtx sync.RWMutex
	task map[int]Task
}

// Repository - интерфейс с методом создания задачи
type Repository interface {
	CreateTask(ctx context.Context, task Task) (int, error) // Создание задачи
	GetTask(ctx context.Context, id int) (Task, error)
}

// NewRepository - создание нового экземпляра репозитория с подключением к PostgreSQL
func NewRepository() (Repository) {
	db:=repository{}
	db.task=map[int]Task{}
	return &db
}

// CreateTask - вставка новой задачи в таблицу tasks
func (r *repository) CreateTask(ctx context.Context, task Task) (int, error) {
	id:=0
	r.mtx.Lock()
	defer r.mtx.Unlock()

	for id+1>id{
		_, ok:=r.task[id]
		if !ok {
			task.ID=id
			r.task[id]=task
			return id, nil}
		id++
	}
	
	return -1, errors.New("free ID not found")
}

// GetTask - вставка новой задачи в таблицу tasksCreate
func (r *repository) GetTask(ctx context.Context, id int) (Task, error) {
	return Task{},nil
}
