package models

import (
	"time"
)

type Reaction struct {
	// The unique numeric id of this reaction
	ID int64 `xorm:"autoincr not null unique pk" json:"-" param:"reaction"`

	// The user who reacted
	User   *User `xorm:"-" json:"user" valid:"-"`
	UserID int64 `xorm:"bigint not null INDEX" json:"-"`

	// The id of the entity you're reacting to
	EntityID int64 `xorm:"bigint not null INDEX" json:"-" param:"entityid"`
	// The entity kind which you're reacting to. Can be 0 for task, 1 for comment.
	EntityKindString string `xorm:"-" json:"-" param:"entitykind"`

	// The actual reaction. This can be any valid utf character or text, up to a length of 20.
	Value string `xorm:"varchar(20) not null INDEX" json:"value" valid:"required"`

	// A timestamp when this reaction was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
}
type ReactionMap map[string][]*User
