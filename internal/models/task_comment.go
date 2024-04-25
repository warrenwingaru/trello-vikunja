package models

import (
	"time"
)

type TaskComment struct {
	ID       int64  `xorm:"autoincr pk unique not null" json:"id" param:"commentid"`
	Comment  string `xorm:"text not null" json:"comment"`
	AuthorID int64  `xorm:"not null" json:"-"`
	Author   *User  `xorm:"-" json:"author"`
	TaskID   int64  `xorm:"not null" json:"-" param:"task"`

	Reactions ReactionMap `xorm:"-" json:"reactions"`

	Created time.Time `xorm:"created" json:"created"`
	Updated time.Time `xorm:"updated" json:"updated"`
}
