package learn

import (
	"errors"
	"net/url"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalidID          = errors.New("invalid id")
	ErrEmptyName          = errors.New("name cannot be empty")
	ErrNameTooLong        = errors.New("name is too long")
	ErrDescriptionTooLong = errors.New("description is too long")
	ErrEmptyNotation      = errors.New("notation cannot be empty")
	ErrNotationTooLong    = errors.New("notation is too long")
	ErrInvalidVideoURL    = errors.New("video URL must be a valid http(s) URL")

	ErrEventNotFound     = errors.New("event not found")
	ErrCaseNotFound      = errors.New("case not found")
	ErrAlgorithmNotFound = errors.New("algorithm not found")
)

const (
	MaxNameLength        = 120
	MaxDescriptionLength = 2000
	MaxNotationLength    = 500
	MaxVideoURLLength    = 500
)

func ValidateEvent(e Event) error {
	if e.ID == "" {
		return ErrInvalidID
	}
	return validateNameAndDescription(e.Name, e.Description)
}

func ValidateCase(c Case) error {
	if c.ID == "" || c.EventID == "" {
		return ErrInvalidID
	}
	return validateNameAndDescription(c.Name, c.Description)
}

func ValidateAlgorithm(a Algorithm) error {
	if a.ID == "" || a.CaseID == "" {
		return ErrInvalidID
	}
	if err := validateNameAndDescription(a.Name, a.Description); err != nil {
		return err
	}
	if strings.TrimSpace(a.Notation) == "" {
		return ErrEmptyNotation
	}
	if utf8.RuneCountInString(a.Notation) > MaxNotationLength {
		return ErrNotationTooLong
	}
	return ValidateVideoURL(a.VideoURL)
}

func validateNameAndDescription(name, description string) error {
	if strings.TrimSpace(name) == "" {
		return ErrEmptyName
	}
	if utf8.RuneCountInString(name) > MaxNameLength {
		return ErrNameTooLong
	}
	if utf8.RuneCountInString(description) > MaxDescriptionLength {
		return ErrDescriptionTooLong
	}
	return nil
}

// ValidateVideoURL: a videó opcionális, de ha van, akkor csak abszolút http(s)
// URL lehet. Ez nem csak formai kérdés — az érték a kliensen egy iframe/anchor
// forrásába kerül, és így nem lehet belőle `javascript:` vagy `data:` séma.
func ValidateVideoURL(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	if utf8.RuneCountInString(raw) > MaxVideoURLLength {
		return ErrInvalidVideoURL
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return ErrInvalidVideoURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidVideoURL
	}
	if parsed.Host == "" {
		return ErrInvalidVideoURL
	}
	return nil
}

// IsValidationError: a handler ebből dönti el, hogy 400-at vagy 500-at ad.
func IsValidationError(err error) bool {
	switch {
	case errors.Is(err, ErrInvalidID),
		errors.Is(err, ErrEmptyName),
		errors.Is(err, ErrNameTooLong),
		errors.Is(err, ErrDescriptionTooLong),
		errors.Is(err, ErrEmptyNotation),
		errors.Is(err, ErrNotationTooLong),
		errors.Is(err, ErrInvalidVideoURL):
		return true
	}
	return false
}
