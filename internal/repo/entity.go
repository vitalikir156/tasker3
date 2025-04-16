package repo

import "time"

// Task - структура, соответствующая таблице tasks.
type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	UID         int       `json:"uid"`
	Created     time.Time `json:"createdAt"`
}

type User struct {
	ID        int    `json:"id"`
	Name      string `json:"uname"`
	Taskread  *bool
	Taskwrite *bool
	Userread  *bool
	Userwrite *bool
}
