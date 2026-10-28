package certificate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"speedcubinghungary/internal/types"

	"github.com/google/uuid"
)

var ErrValidation = errors.New("invalid certificate template")

type validationError string

func (e validationError) Error() string        { return string(e) }
func (e validationError) Is(target error) bool { return target == ErrValidation }

const (
	MaxNameLength        = 120
	MaxDescriptionLength = 500
	MaxDocumentBytes     = 1 << 20
	MaxElements          = 200
	DocumentSchema       = 1
)

var (
	ErrInvalidTemplateID error = validationError("invalid template id")
	ErrEmptyName         error = validationError("template name cannot be empty")
	ErrNameTooLong       error = validationError(fmt.Sprintf("template name must be at most %d characters", MaxNameLength))
	ErrDescriptionLong   error = validationError(fmt.Sprintf("description must be at most %d characters", MaxDescriptionLength))
	ErrInvalidKind       error = validationError("template kind must be participation or placement")
	ErrInvalidDocument   error = validationError("template document must be an object with schemaVersion, page and elements")
	ErrDocumentTooLarge  error = validationError(fmt.Sprintf("template document must be at most %d KiB", MaxDocumentBytes>>10))
	ErrTooManyElements   error = validationError(fmt.Sprintf("a template can have at most %d elements", MaxElements))
	ErrInvalidAssetID    error = validationError("invalid asset id")

	ErrTemplateNotFound = errors.New("certificate template not found")
	ErrVersionConflict  = errors.New("the template was saved by someone else in the meantime")
	ErrAssetNotFound    = errors.New("certificate asset not found")
	ErrAssetInUse       = errors.New("the image is used by a template")
)

func validID(id string, invalid error) error {
	if _, err := uuid.Parse(id); err != nil {
		return invalid
	}
	return nil
}

func validKind(kind types.CertificateTemplateKind) bool {
	return kind == types.CERTIFICATE_PARTICIPATION || kind == types.CERTIFICATE_PLACEMENT
}

func normalizeInput(input TemplateInput) (TemplateInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	switch {
	case input.Name == "":
		return input, ErrEmptyName
	case utf8.RuneCountInString(input.Name) > MaxNameLength:
		return input, ErrNameTooLong
	case utf8.RuneCountInString(input.Description) > MaxDescriptionLength:
		return input, ErrDescriptionLong
	case !validKind(input.Kind):
		return input, ErrInvalidKind
	}

	document, err := validDocument(input.Document)
	if err != nil {
		return input, err
	}
	input.Document = document
	return input, nil
}

func validDocument(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) > MaxDocumentBytes {
		return nil, ErrDocumentTooLarge
	}

	var document struct {
		SchemaVersion int               `json:"schemaVersion"`
		Page          *json.RawMessage  `json:"page"`
		Elements      []json.RawMessage `json:"elements"`
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(&document); err != nil || decoder.More() {
		return nil, ErrInvalidDocument
	}
	if document.SchemaVersion != DocumentSchema || document.Page == nil || document.Elements == nil {
		return nil, ErrInvalidDocument
	}
	if len(document.Elements) > MaxElements {
		return nil, ErrTooManyElements
	}

	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return nil, ErrInvalidDocument
	}
	return compact.Bytes(), nil
}
