package models

import "time"

type Label struct {
	ID          int64     `json:"id"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Description string    `json:"description"`
	Title       string    `json:"title"`
	HexColor    string    `json:"hex_color"`
}
