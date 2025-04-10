package repo

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

var ErrTaskNotFound = errors.New("task with ID not found")
var ErrFreeIDNotFound = errors.New("free ID not found")

type repository struct {
	mtx sync.RWMutex
	task map[int]Task
}

type Repository interface {
	CreateTask(ctx context.Context, task Task) (int, error)
	GetTask(ctx context.Context, id int) (Task, error)
	GetTasks(ctx context.Context) ([]Task, error)
	UpdateTask(ctx context.Context, task Task) (error)
	DeleteTask(ctx context.Context, id int) (error)
}

func NewRepository() (Repository) {
	db:=repository{}
	db.task=map[int]Task{}
	return &db
}


func (r *repository) CreateTask(ctx context.Context, task Task) (int, error) {
	id:=0
	r.mtx.Lock()
	defer r.mtx.Unlock()

	for id+1>id{
		_, ok:=r.task[id]
		if !ok {
			task.Status="new"
			r.task[id]=task
			return id, nil}
		id++
	}
	
	return -1, ErrFreeIDNotFound
}


func (r *repository) GetTask(ctx context.Context, id int) (Task, error) {
	r.mtx.RLock()
	task, ok:=r.task[id]
	r.mtx.RUnlock()
	if ok {
		task.ID=id
		return task,nil
	}

	return Task{}, ErrTaskNotFound
}

func (r *repository) GetTasks(ctx context.Context) ([]Task, error) {
	tasks:=make([]Task, 0, len(r.task))
	r.mtx.RLock()
	for k, v:=range r.task{
		v.ID=k
		tasks=append(tasks, v)
	}
	r.mtx.RUnlock()
	return tasks,nil
}

func (r *repository) UpdateTask(ctx context.Context, task Task) (error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	_, ok:=r.task[task.ID]
	if !ok{return ErrTaskNotFound}

	r.task[task.ID]=task

	return nil
}

func (r *repository) DeleteTask(ctx context.Context, id int) (error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	_, ok:=r.task[id]
	if !ok{return ErrTaskNotFound}

	delete(r.task, id)

	return nil
}