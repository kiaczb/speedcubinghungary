package post

import "errors"

var (
	ErrInvalidID    = errors.New("invalid post ID")
	ErrInvalidTag	= errors.New("invalid tag")
	ErrEmptyTitle   = errors.New("title cannot be empty")
	ErrShortContent = errors.New("content must be at least 20 characters")
)

func ValidatePost(p Post) error {
	if p.ID == "" {
		return ErrInvalidID
	}
	if p.Tag == "" {
		return ErrInvalidTag
	}
	if p.Title == "" {
		return ErrEmptyTitle
	}
	if len(p.Content) < 20 {
		return ErrShortContent
	}
	return nil
}
