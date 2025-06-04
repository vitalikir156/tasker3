package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vitalikir156/tasker3/internal/dto"
	"github.com/vitalikir156/tasker3/internal/repo"
	"github.com/vitalikir156/tasker3/internal/repo/mocks"
	"go.uber.org/zap"
)

func boolPtr(b bool) *bool {
	return &b
}

func setupTest() (*fiber.App, *mocks.Repository, *zap.SugaredLogger) {
	app := fiber.New()
	mockRepo := new(mocks.Repository)
	logger := zap.NewNop().Sugar()
	return app, mockRepo, logger
}

func TestCreateTaskGood(t *testing.T) {
	app, mockRepo, logger := setupTest()

	s := NewService(mockRepo, logger)

	app.Post("/tasks", s.CreateTask)
	
	task := TaskRequest{
			Title:       "Test Task",
			Description: "Test Description",
			Userrequest: Userrequest{UserID: "1"},
		}
		body, _ := json.Marshal(task)
		uid, _ := strconv.Atoi(task.UserID)
		// Ожидаем, что вызов метода `CreateTask` в репозитории вернёт ID = 1
		mockRepo.On("CreateTask", mock.Anything, repo.Task{
			Title:       task.Title,
			Description: task.Description,
			UID: uid,
		}).Return(1, nil).Once()

		// Отправляем запрос
		req, err := http.NewRequest("POST", "/tasks", bytes.NewReader(body))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		
		// Выполняем запрос
		resp, err := app.Test(req)
		
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Проверяем ответ
		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "success", response.Status)
		
		// Проверяем вызов мок-методов
		mockRepo.AssertExpectations(t)
}
func TestCreateTaskValidation(t *testing.T) {
	app, mockRepo, logger := setupTest()

	s := NewService(mockRepo, logger)

	app.Post("/tasks", s.CreateTask)
	body := []byte(`{}`) // Пустое тело, `title` обязателен

	req, err := http.NewRequest("POST", "/tasks", bytes.NewReader(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response dto.Response
	
	json.NewDecoder(resp.Body).Decode(&response)
	assert.Equal(t, "error", response.Status)
}
func TestCreateTaskRepoFail(t *testing.T) {
	app, mockRepo, logger := setupTest()

	s := NewService(mockRepo, logger)

	app.Post("/tasks", s.CreateTask)

	task := TaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
		Userrequest: Userrequest{UserID: "1"},
	}
	body, _ := json.Marshal(task)

	// Ожидаем ошибку при вставке в БД
	mockRepo.On("CreateTask", mock.Anything, repo.Task{
		Title:       task.Title,
		Description: task.Description,
		UID: 1,
	}).Return(0, errors.New("DB error")).Once()

	req, err := http.NewRequest("POST", "/tasks", bytes.NewReader(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	var response dto.Response
	
	json.NewDecoder(resp.Body).Decode(&response)
	assert.Equal(t, "error", response.Status)
	mockRepo.AssertExpectations(t)
}
func TestGetTaskGood(t *testing.T) {
	app, mockRepo, logger := setupTest()
	s := NewService(mockRepo, logger)
	app.Get("/tasks/:id", s.GetTask)

	reqBody := `{"userId": "1"}`
		mockUser := repo.User{
			ID:       1,
			Taskread: boolPtr(true),
		}
		mockTask := repo.Task{
			ID:     1,
			Title:  "Test Task",
			UID:    1,
		}

		mockRepo.On("GetUser", mock.Anything, 1).Return(mockUser, nil)
		mockRepo.On("GetTask", mock.Anything, 1).Return(mockTask, nil)

		req, err := http.NewRequest("GET", "/tasks/1", bytes.NewReader([]byte(reqBody)))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var task repo.Task
		json.NewDecoder(resp.Body).Decode(&task)
		assert.Equal(t, "Test Task", task.Title)
		mockRepo.AssertExpectations(t)
}

func TestGetTaskAccessDenied(t *testing.T) {
	app, mockRepo, logger := setupTest()
	s := NewService(mockRepo, logger)
	app.Get("/tasks/:id", s.GetTask)

	reqBody := `{"userId": "1"}`
	mockUser := repo.User{
		ID:       1,
		Taskread: boolPtr(false),
	}
	mockTask := repo.Task{
		ID:     1,
		Title:  "Test Task",
		UID:    2, // Другой пользователь
	}
	mockRepo.On("GetUser", mock.Anything, 1).Return(mockUser, nil)
	mockRepo.On("GetTask", mock.Anything, 1).Return(mockTask, nil)
	req, err := http.NewRequest("GET", "/tasks/1", bytes.NewReader([]byte(reqBody)))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	mockRepo.AssertExpectations(t)
}

func TestGetTaskNotFound(t *testing.T) {
	app, mockRepo, logger := setupTest()
	s := NewService(mockRepo, logger)
	app.Get("/tasks/:id", s.GetTask)

	reqBody := `{"userId": "1"}`
		mockUser := repo.User{
			ID:       1,
			Taskread: boolPtr(true),
		}

		mockRepo.On("GetUser", mock.Anything, 1).Return(mockUser, nil)
		mockRepo.On("GetTask", mock.Anything, 1).Return(repo.Task{}, pgx.ErrNoRows)

		req, err := http.NewRequest("GET", "/tasks/1", bytes.NewReader([]byte(reqBody)))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		mockRepo.AssertExpectations(t)
}



func TestGetTasksGoodWithPerm(t *testing.T) {
	app, mockRepo, logger := setupTest()
	s := NewService(mockRepo, logger)
	app.Get("/tasks", s.GetTasks)

	reqBody := `{"userId": "1"}`
		mockUser := repo.User{
			ID:       1,
			Taskread: boolPtr(true),
		}
		mockTasks := []repo.Task{
			{ID: 1, Title: "Task 1"},
			{ID: 2, Title: "Task 2"},
		}

		mockRepo.On("GetUser", mock.Anything, 1).Return(mockUser, nil)
		mockRepo.On("GetTasks", mock.Anything).Return(mockTasks, nil)

		req, err := http.NewRequest("GET", "/tasks", bytes.NewReader([]byte(reqBody)))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var tasks []repo.Task
		json.NewDecoder(resp.Body).Decode(&tasks)
		assert.Len(t, tasks, 2)
		mockRepo.AssertExpectations(t)

}

func TestGetTasksGoodWithoutPerm(t *testing.T) {
	app, mockRepo, logger := setupTest()
	s := NewService(mockRepo, logger)
	app.Get("/tasks", s.GetTasks)

	reqBody := `{"userId": "1"}`
	mockUser := repo.User{
		ID:       1,
		Taskread: boolPtr(false),
	}
	mockTasks := []repo.Task{
		{ID: 1, Title: "Task 1", UID: 1},
	}

	mockRepo.On("GetUser", mock.Anything, 1).Return(mockUser, nil)
	mockRepo.On("GetTasksOverUID", mock.Anything, 1).Return(mockTasks, nil)

	req, err := http.NewRequest("GET", "/tasks", bytes.NewReader([]byte(reqBody)))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var tasks []repo.Task
	json.NewDecoder(resp.Body).Decode(&tasks)
	assert.Len(t, tasks, 1)
	mockRepo.AssertExpectations(t)

}

func TestGetTasksBadUserNotFound(t *testing.T) {
	app, mockRepo, logger := setupTest()
	s := NewService(mockRepo, logger)
	app.Get("/tasks", s.GetTasks)

	reqBody := `{"userId": "1"}`
		mockRepo.On("GetUser", mock.Anything, 1).Return(repo.User{}, pgx.ErrNoRows)

		req, err := http.NewRequest("GET", "/tasks", bytes.NewReader([]byte(reqBody)))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		mockRepo.AssertExpectations(t)
}

func TestUpdateTaskGood(t *testing.T) {
	app, mockRepo, logger := setupTest()
	s := NewService(mockRepo, logger)
	app.Put("/tasks/:id", s.UpdateTask)

	reqBody := `{
		"title": "Updated Task",
		"description": "Updated Desc",
		"status": "in_progress",
		"uid": "1",
		"userId": "1"
	}`
	mockUser := repo.User{
		ID:        1,
		Taskwrite: boolPtr(true),
	}
	mockTask := repo.Task{
		ID:  1,
		UID: 1,
	}

	mockRepo.On("GetUser", mock.Anything, 1).Return(mockUser, nil)
	mockRepo.On("GetTask", mock.Anything, 1).Return(mockTask, nil)
	mockRepo.On("UpdateTask", mock.Anything, repo.Task{
		ID:          1,
		Title:       "Updated Task",
		Description: "Updated Desc",
		Status:      "in_progress",
		UID:         1,
	}).Return(nil)

	req, err := http.NewRequest("PUT", "/tasks/1", bytes.NewReader([]byte(reqBody)))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockRepo.AssertExpectations(t)

}

func TestUpdateTaskAccessDenied(t *testing.T) {
	app, mockRepo, logger := setupTest()
	s := NewService(mockRepo, logger)
	app.Put("/tasks/:id", s.UpdateTask)

	reqBody := `{
		"title": "Updated Task",
		"userId": "1",
		"uid": "1"
	}`
	mockUser := repo.User{
		ID:        1,
		Taskwrite: boolPtr(false),
	}
	mockTask := repo.Task{
		ID:  1,
		UID: 2, // Другой пользователь
	}

	mockRepo.On("GetUser", mock.Anything, 1).Return(mockUser, nil)
	mockRepo.On("GetTask", mock.Anything, 1).Return(mockTask, nil)

	req, err := http.NewRequest("PUT", "/tasks/1", bytes.NewReader([]byte(reqBody)))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	mockRepo.AssertExpectations(t)
}

func TestDeleteTaskGood(t *testing.T) {
	app, mockRepo, logger := setupTest()
	s := NewService(mockRepo, logger)
	app.Delete("/tasks/:id", s.DeleteTask)

	reqBody := `{"userId": "1"}`
		mockUser := repo.User{
			ID:        1,
			Taskwrite: boolPtr(true),
		}
		mockTask := repo.Task{
			ID:  1,
			UID: 1,
		}

		mockRepo.On("GetUser", mock.Anything, 1).Return(mockUser, nil)
		mockRepo.On("GetTask", mock.Anything, 1).Return(mockTask, nil)
		mockRepo.On("DeleteTask", mock.Anything, 1).Return(nil)

		req, err := http.NewRequest("DELETE", "/tasks/1", bytes.NewReader([]byte(reqBody)))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		mockRepo.AssertExpectations(t)
}


func TestDeleteTaskAccessDenied(t *testing.T) {
	app, mockRepo, logger := setupTest()
	s := NewService(mockRepo, logger)
	app.Delete("/tasks/:id", s.DeleteTask)
	reqBody := `{"userId": "1"}`
		mockUser := repo.User{
			ID:        1,
			Taskwrite: boolPtr(false),
		}
		mockTask := repo.Task{
			ID:  1,
			UID: 2, // Другой пользователь
		}

		mockRepo.On("GetUser", mock.Anything, 1).Return(mockUser, nil)
		mockRepo.On("GetTask", mock.Anything, 1).Return(mockTask, nil)

		req, err := http.NewRequest("DELETE", "/tasks/1", bytes.NewReader([]byte(reqBody)))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		mockRepo.AssertExpectations(t)
}

