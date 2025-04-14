package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/vitalikir156/tasker3/internal/dto"
	"github.com/vitalikir156/tasker3/internal/repo"
	"github.com/vitalikir156/tasker3/pkg/validator"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Слой бизнес-логики. Тут должна быть основная логика сервиса

// Service - интерфейс для бизнес-логики
type Service interface {
	CreateTask(ctx *fiber.Ctx) error
	GetTask(ctx *fiber.Ctx) error
	GetTasks(ctx *fiber.Ctx) error
	UpdateTask(ctx *fiber.Ctx) error
	DeleteTask(ctx *fiber.Ctx) error
}

type service struct {
	repo repo.Repository
	log  *zap.SugaredLogger
}

// NewService - конструктор сервиса
func NewService(repo repo.Repository, logger *zap.SugaredLogger) Service {
	return &service{
		repo: repo,
		log:  logger,
	}
}


func (s *service) GetTask(ctx *fiber.Ctx) error { //GetTask
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		s.log.Error("Invalid request", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "Invalid request")
	}
	task, err := s.repo.GetTask(ctx.Context(), id)
	if err != nil {
		s.log.Info("Failed to get task", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows){
			return ctx.Status(fiber.StatusNotFound).SendString("task with ID not found")
		}
		return dto.InternalServerError(ctx)
	}
	return ctx.Status(fiber.StatusOK).JSON(task)
}

func (s *service) GetTasks(ctx *fiber.Ctx) error { //GetTask
	tasks, err := s.repo.GetTasks(ctx.Context())
	if err != nil {
		s.log.Error("Failed to get tasks", zap.Error(err))
		return dto.InternalServerError(ctx)
	}
	return ctx.Status(fiber.StatusOK).JSON(tasks)
}


func (s *service) CreateTask(ctx *fiber.Ctx) error {
	var req TaskRequest

	// Десериализация JSON-запроса
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		s.log.Error("Invalid request body", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "Invalid request body")
	}

	// Валидация входных данных
	if vErr := validator.Validate(ctx.Context(), req); vErr != nil {
		return dto.BadResponseError(ctx, dto.FieldIncorrect, vErr.Error())
	}
	if len(req.UID)==0{s.log.Error("Empty UID")
	return dto.BadResponseError(ctx, dto.FieldBadFormat, "Invalid request body")}
	uid, err:=strconv.Atoi(req.UID)
	if err != nil {
		s.log.Error("Failed to convert UID", zap.Error(err))
		return dto.InternalServerError(ctx)
	}
	// Вставка задачи в БД через репозиторий
	task := repo.Task{
		Title:       req.Title,
		Description: req.Description,
		Status: req.Status,
		UID: uid,
	}
	taskID, err := s.repo.CreateTask(ctx.Context(), task)
	if err != nil {
		s.log.Error("Failed to insert task", zap.Error(err))
		return dto.InternalServerError(ctx)
	}

	// Формирование ответа
	response := dto.Response{
		Status: "success",
		Data:   map[string]int{"task_id": taskID},
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *service) UpdateTask(ctx *fiber.Ctx) error {
	var req TaskRequest
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		s.log.Error("Invalid request", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "Invalid request")
	}
	// Десериализация JSON-запроса
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		s.log.Error("Invalid request body", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "Invalid request body")
	}

	// Валидация входных данных
	if vErr := validator.Validate(ctx.Context(), req); vErr != nil {
		return dto.BadResponseError(ctx, dto.FieldIncorrect, vErr.Error())
	}
	uid, err:=strconv.Atoi(req.UID)
	if err != nil {
		s.log.Error("Failed to convert UID", zap.Error(err))
		return dto.InternalServerError(ctx)
	}
	// Вставка задачи в БД через репозиторий
	task := repo.Task{
		Title:       req.Title,
		Description: req.Description,
		Status: req.Status,
		ID: id,
		UID: uid,
	}
	err = s.repo.UpdateTask(ctx.Context(), task)
	if err != nil {
		s.log.Error("Failed to update task", zap.Error(err))
		if errors.Is(err, repo.ErrTaskNotFound){
			return ctx.Status(fiber.StatusNotFound).SendString("task with ID not found")
		}
		return dto.InternalServerError(ctx)
	}

	// Формирование ответа
	response := dto.Response{
		Status: "success",
	}
	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *service) DeleteTask(ctx *fiber.Ctx) error { //GetTask
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		s.log.Error("Invalid request", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "Invalid request")
	}
	fmt.Println(id, err)
	err = s.repo.DeleteTask(ctx.Context(), id)
	if err != nil {
		s.log.Error("Failed to delete task", zap.Error(err))
		if errors.Is(err, repo.ErrTaskNotFound){
			return ctx.Status(fiber.StatusNotFound).SendString("task with ID not found")
		}
		return dto.InternalServerError(ctx)
	}
	response := dto.Response{
		Status: "success",
	}
	return ctx.Status(fiber.StatusOK).JSON(response)
}