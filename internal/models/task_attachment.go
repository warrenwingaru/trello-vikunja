package models

import "time"

type TaskAttachment struct {
	ID     int64 `json:"id"`
	TaskID int64 `json:"task_id"`
	FileID int64 `json:"-"`

	CreatedByID int64 `json:"-"`
	CreatedBy   *User `json:"created_by"`

	File    *File     `json:"file"`
	Created time.Time `json:"created"`
}
