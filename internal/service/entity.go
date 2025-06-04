package service

// TaskRequest - структура, представляющая тело запроса.
type TaskRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Userrequest
}
type Userrequest struct {
	UserID string `json:"userId" validate:"required"`
}
