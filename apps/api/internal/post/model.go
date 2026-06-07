package post

import (
	"speedcubinghungary/internal/types"
	"time"
)


type Post struct {
	ID        string    `json:"id"`
	Tag       types.PostTag `json:"tag"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}