package post

import (
	"errors"
	"speedcubinghungary/internal/types"
)

var (
	ErrInvalidID    = errors.New("invalid post ID")
	ErrInvalidTag	= errors.New("invalid tag")
	ErrEmptyTitle   = errors.New("title cannot be empty")
	ErrShortContent = errors.New("content must be at least 20 characters")
)

// isValidTag: a post_tag DB enum megengedett értékei.
func isValidTag(t types.PostTag) bool {
	switch t {
	case types.TAG_EVENT, types.TAG_COMPETITION, types.TAG_NEWS,
		types.TAG_ANNOUNCEMENT, types.TAG_REPORT:
		return true
	}
	return false
}

func ValidatePost(p Post) error {
	if p.ID == "" {
		return ErrInvalidID
	}
	// A tag-et a post_tag DB enum ellen is ellenőrizzük: enélkül egy hibás érték
	// csak az INSERT-nél bukna el, 500-as hibaként a 400 helyett.
	if !isValidTag(p.Tag) {
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
