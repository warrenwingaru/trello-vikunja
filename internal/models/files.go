package models

import (
	"github.com/spf13/afero"
	"time"
)

type File struct {
	ID   int64  `xorm:"bigint autoincr not null unique pk" json:"id"`
	Name string `xorm:"text not null" json:"name"`
	Mime string `xorm:"text null" json:"mime"`
	Size uint64 `xorm:"bigint not null" json:"size"`

	Created     time.Time `xorm:"created" json:"created"`
	CreatedByID int64     `xorm:"bigint not null" json:"-"`

	File afero.File `xorm:"-" json:"-"`
	// This ReadCloser is only used for migration purposes. Use with care!
	// There is currentlc no better way of doing this.
	FileContent []byte `xorm:"-" json:"-"`
}
