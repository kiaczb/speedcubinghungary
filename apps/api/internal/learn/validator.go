package learn

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	ErrValidation = errors.New("invalid learning content")
	ErrConflict   = errors.New("conflicting learning content")
)

type validationError string

func (e validationError) Error() string        { return string(e) }
func (e validationError) Is(target error) bool { return target == ErrValidation }

type conflictError string

func (e conflictError) Error() string        { return string(e) }
func (e conflictError) Is(target error) bool { return target == ErrConflict }

const (
	MaxDepth               = 32
	MaxTitleLength         = 100
	MaxSlugLength          = 64
	MaxDescriptionLength   = 1000
	MaxNotationLength      = 500
	MaxAlgorithmNameLength = 100
	MaxNotesLength         = 2000
	MaxOrderLength         = 1000
	MaxImageURLLength      = 2000
)

var (
	ErrInvalidNodeID        error = validationError("invalid node id")
	ErrInvalidParentID      error = validationError("invalid parent id")
	ErrInvalidAlgorithmID   error = validationError("invalid algorithm id")
	ErrInvalidPath          error = validationError("invalid path")
	ErrEmptyTitle           error = validationError("title cannot be empty")
	ErrTitleTooLong         error = validationError(fmt.Sprintf("title must be at most %d characters", MaxTitleLength))
	ErrInvalidSlug          error = validationError("no slug can be derived — give one using latin letters or digits")
	ErrDescriptionTooLong   error = validationError(fmt.Sprintf("description must be at most %d characters", MaxDescriptionLength))
	ErrInvalidIcon          error = validationError("icon must be a cubing icon class such as event-333 or unofficial-fto")
	ErrInvalidImageURL      error = validationError(fmt.Sprintf("image url must be an https:// address of at most %d characters", MaxImageURLLength))
	ErrEmptyNotation        error = validationError("algorithm notation cannot be empty")
	ErrNotationTooLong      error = validationError(fmt.Sprintf("algorithm notation must be at most %d characters", MaxNotationLength))
	ErrAlgorithmNameTooLong error = validationError(fmt.Sprintf("algorithm name must be at most %d characters", MaxAlgorithmNameLength))
	ErrNotesTooLong         error = validationError(fmt.Sprintf("notes must be at most %d characters", MaxNotesLength))
	ErrInvalidOrder         error = validationError("order must list valid ids, each exactly once")
	ErrOrderTooLong         error = validationError(fmt.Sprintf("order can list at most %d ids", MaxOrderLength))

	ErrSlugTaken           error = conflictError("another node on this level already uses this slug")
	ErrParentHasAlgorithms error = conflictError("a node that has algorithms cannot have child nodes")
	ErrNodeHasChildren     error = conflictError("a node that has child nodes cannot have algorithms")
	ErrMoveIntoOwnSubtree  error = conflictError("a node cannot be moved under itself or its own descendants")
	ErrTooDeep             error = conflictError(fmt.Sprintf("the hierarchy can be at most %d levels deep", MaxDepth))
	ErrOrderMismatch       error = conflictError("the order does not match the current items — reload and try again")

	ErrNodeNotFound      = errors.New("node not found")
	ErrParentNotFound    = errors.New("parent node not found")
	ErrAlgorithmNotFound = errors.New("algorithm not found")
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var iconPattern = regexp.MustCompile(`^(?:event|unofficial)-[a-z0-9_]{1,48}$`)

var transliterations = map[rune]string{
	'á': "a", 'à': "a", 'â': "a", 'ä': "a", 'ã': "a", 'å': "a",
	'é': "e", 'è': "e", 'ê': "e", 'ë': "e",
	'í': "i", 'ì': "i", 'î': "i", 'ï': "i",
	'ó': "o", 'ò': "o", 'ô': "o", 'ö': "o", 'õ': "o", 'ő': "o", 'ø': "o",
	'ú': "u", 'ù': "u", 'û': "u", 'ü': "u", 'ű': "u",
	'ý': "y", 'ÿ': "y", 'ñ': "n", 'ç': "c", 'ß': "ss",
}

func Slugify(text string) string {
	var b strings.Builder
	separate := false

	emit := func(r rune) {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			separate = true
			return
		}
		if separate && b.Len() > 0 {
			b.WriteByte('-')
		}
		separate = false
		b.WriteRune(r)
	}

	for _, r := range strings.ToLower(text) {
		if latin, ok := transliterations[r]; ok {
			for _, l := range latin {
				emit(l)
			}
			continue
		}
		emit(r)
	}

	slug := b.String()
	if len(slug) > MaxSlugLength {
		slug = strings.TrimRight(slug[:MaxSlugLength], "-")
	}
	return slug
}

func canonicalID(id string, invalid error) (string, error) {
	parsed, err := uuid.Parse(id)
	if err != nil || len(id) != 36 {
		return "", invalid
	}
	return parsed.String(), nil
}

func validatePath(path []string) error {
	if len(path) > MaxDepth {
		return ErrInvalidPath
	}
	for _, segment := range path {
		if len(segment) > MaxSlugLength || !slugPattern.MatchString(segment) {
			return ErrInvalidPath
		}
	}
	return nil
}

func normalizeNode(in NodeInput) (NodeInput, error) {
	if in.ParentID != nil {
		parentID, err := canonicalID(*in.ParentID, ErrInvalidParentID)
		if err != nil {
			return NodeInput{}, err
		}
		in.ParentID = &parentID
	}

	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	in.Icon = strings.TrimSpace(in.Icon)
	in.ImageURL = strings.TrimSpace(in.ImageURL)

	source := strings.TrimSpace(in.Slug)
	if source == "" {
		source = in.Title
	}
	in.Slug = Slugify(source)

	return in, ValidateNode(in)
}

func ValidateNode(in NodeInput) error {
	if in.Title == "" {
		return ErrEmptyTitle
	}
	if utf8.RuneCountInString(in.Title) > MaxTitleLength {
		return ErrTitleTooLong
	}
	if !slugPattern.MatchString(in.Slug) || len(in.Slug) > MaxSlugLength {
		return ErrInvalidSlug
	}
	if utf8.RuneCountInString(in.Description) > MaxDescriptionLength {
		return ErrDescriptionTooLong
	}
	if in.Icon != "" && !iconPattern.MatchString(in.Icon) {
		return ErrInvalidIcon
	}
	if in.ImageURL != "" && !isImageURL(in.ImageURL) {
		return ErrInvalidImageURL
	}
	return nil
}

func isImageURL(raw string) bool {
	if len(raw) > MaxImageURLLength || strings.ContainsAny(raw, " \t\r\n\"'<>") {
		return false
	}
	parsed, err := url.Parse(raw)
	return err == nil && parsed.Scheme == "https" && parsed.Host != "" && parsed.User == nil
}

func normalizeAlgorithm(in AlgorithmInput) (AlgorithmInput, error) {
	in.Notation = strings.Join(strings.Fields(in.Notation), " ")
	in.Name = strings.TrimSpace(in.Name)
	in.Notes = strings.TrimSpace(in.Notes)

	return in, ValidateAlgorithm(in)
}

func ValidateAlgorithm(in AlgorithmInput) error {
	if in.Notation == "" {
		return ErrEmptyNotation
	}
	if utf8.RuneCountInString(in.Notation) > MaxNotationLength {
		return ErrNotationTooLong
	}
	if utf8.RuneCountInString(in.Name) > MaxAlgorithmNameLength {
		return ErrAlgorithmNameTooLong
	}
	if utf8.RuneCountInString(in.Notes) > MaxNotesLength {
		return ErrNotesTooLong
	}
	return nil
}

func normalizeOrder(ids []string) ([]string, error) {
	if len(ids) > MaxOrderLength {
		return nil, ErrOrderTooLong
	}

	canonical := make([]string, len(ids))
	seen := make(map[string]bool, len(ids))
	for i, id := range ids {
		c, err := canonicalID(id, ErrInvalidOrder)
		if err != nil {
			return nil, err
		}
		if seen[c] {
			return nil, ErrInvalidOrder
		}
		seen[c] = true
		canonical[i] = c
	}
	return canonical, nil
}
