package service

// TaskRequest - структура, представляющая тело запроса
type TaskRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	Status      string `json:"status"`
	UID			string `json:"uid"` //user id in task
	User_ID 	string `json:"user_id"` //user id in query
}
