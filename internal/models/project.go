package models

import (
	"os/user"
	"time"
)

type Project struct {
	// The unique, numeric id of this project.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"project"`
	// The title of the project. You'll see this in the overview.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	// The description of the project.
	Description string `xorm:"longtext null" json:"description"`
	// The unique project short identifier. Used to build task identifiers.
	Identifier string `xorm:"varchar(10) null" json:"identifier" valid:"runelength(0|10)" minLength:"0" maxLength:"10"`
	// The hex color of this project
	HexColor string `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|7)" maxLength:"7"`

	OwnerID         int64    `xorm:"bigint INDEX not null" json:"-"`
	ParentProjectID int64    `xorm:"bigint INDEX null" json:"parent_project_id"`
	ParentProject   *Project `xorm:"-" json:"-"`

	// The user who created this project.
	Owner *user.User `xorm:"-" json:"owner" valid:"-"`

	// Whether a project is archived.
	IsArchived bool `xorm:"not null default false" json:"is_archived" query:"is_archived"`

	// The id of the file this project has set as background
	BackgroundFileID int64 `xorm:"null" json:"-"`
	// Holds extra information about the background set since some background providers require attribution or similar. If not null, the background can be accessed at /projects/{projectID}/background
	BackgroundInformation interface{} `xorm:"-" json:"background_information"`
	// Contains a very small version of the project background to use as a blurry preview until the actual background is loaded. Check out https://blurha.sh/ to learn how it works.
	BackgroundBlurHash string `xorm:"varchar(50) null" json:"background_blur_hash"`

	// True if a project is a favorite. Favorite projects show up in a separate parent project. This value depends on the user making the call to the api.
	IsFavorite bool `xorm:"-" json:"is_favorite"`

	// The subscription status for the user reading this project. You can only read this property, use the subscription endpoints to modify it.
	// Will only returned when retreiving one project.
	Subscription *interface{} `xorm:"-" json:"subscription,omitempty"`

	// The position this project has when querying all projects. See the tasks.position property on how to use this.
	Position float64 `xorm:"double null" json:"position"`

	Views []*ProjectView `xorm:"-" json:"views"`

	// A timestamp when this project was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this project was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`
}
type ProjectWithTasksAndBuckets struct {
	Project
	ChildProjects []*ProjectWithTasksAndBuckets `xorm:"-" json:"child_projects"`

	// Only used for migration.
	Buckets          []*Bucket       `xorm:"-" json:"buckets"`
	TaskBuckets      []*TaskBucket   `xorm:"-" json:"task_buckets"`
	Positions        []*TaskPosition `xorm:"-" json:"positions"`
	BackgroundFileID int64           `xorm:"null" json:"background_file_id"`
}
