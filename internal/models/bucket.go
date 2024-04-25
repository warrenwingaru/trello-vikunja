package models

import (
	"time"
)

// Bucket represents a kanban bucket
type Bucket struct {
	// The unique, numeric id of this bucket.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"bucket"`
	// The title of this bucket.
	Title string `xorm:"text not null" valid:"required" minLength:"1" json:"title"`
	// The project this bucket belongs to.
	ProjectID int64 `xorm:"-" json:"-" param:"project" json:"project_id"`
	// The project view this bucket belongs to.
	ProjectViewID int64 `xorm:"bigint not null" json:"project_view_id" param:"view"`
	// All tasks which belong to this bucket.
	Tasks             []*Task             `xorm:"-" json:"tasks"`
	TasksWithComments []*TaskWithComments `xorm:"-" json:"-"`

	// How many tasks can be at the same time on this board max
	Limit int64 `xorm:"default 0" json:"limit" minimum:"0" valid:"range(0|9223372036854775807)"`

	// The number of tasks currently in this bucket
	Count int64 `xorm:"-" json:"count"`

	// The position this bucket has when querying all buckets. See the tasks.position property on how to use this.
	Position float64 `xorm:"double null" json:"position"`

	// A timestamp when this bucket was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this bucket was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	// The user who initially created the bucket.
	CreatedBy   *User `xorm:"-" json:"created_by" valid:"-"`
	CreatedByID int64 `xorm:"bigint not null" json:"-"`

	// Including the task collection type so we can use task filters on kanban
}
type TaskBucket struct {
	BucketID      int64 `xorm:"bigint not null index"`
	TaskID        int64 `xorm:"bigint not null index"`
	ProjectViewID int64 `xorm:"bigint not null index"`
}
type TaskPosition struct {
	// The ID of the task this position is for
	TaskID int64 `xorm:"bigint not null index" json:"task_id" param:"task"`
	// The project view this task is related to
	ProjectViewID int64 `xorm:"bigint not null index" json:"project_view_id"`
	// The position of the task - any task project can be sorted as usual by this parameter.
	// When accessing tasks via kanban buckets, this is primarily used to sort them based on a range
	// We're using a float64 here to make it possible to put any task within any two other tasks (by changing the number).
	// You would calculate the new position between two tasks with something like task3.position = (task2.position - task1.position) / 2.
	// A 64-Bit float leaves plenty of room to initially give tasks a position with 2^16 difference to the previous task
	// which also leaves a lot of room for rearranging and sorting later.
	// Positions are always saved per view. They will automatically be set if you request the tasks through a view
	// endpoint, otherwise they will always be 0. To update them, take a look at the Task Position endpoint.
	Position float64 `xorm:"double not null" json:"position"`
}
