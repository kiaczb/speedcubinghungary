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
	// ImagePath a borítókép publikus útvonala (pl. "/uploads/<uuid>.webp"),
	// üres string ha a poszthoz nincs kép. A DB oszlop nullable.
	ImagePath string    `json:"image_path"`
	CreatedAt time.Time `json:"created_at"`
}