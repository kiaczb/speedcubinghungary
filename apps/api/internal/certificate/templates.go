package certificate

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	MaxAssetSize       = 10 << 20
	MaxAssetDimension  = 12000
	maxAssetNameLength = 200
)

var (
	ErrUnsupportedAsset error = validationError("only PNG and JPEG images are supported")
	ErrAssetTooLarge    error = validationError(fmt.Sprintf("image is too large (max %d MiB)", MaxAssetSize>>20))
	ErrAssetDimensions  error = validationError(fmt.Sprintf("image must be at most %d pixels wide and tall", MaxAssetDimension))
)

var assetTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
}

func (s *Service) ListTemplates() ([]Template, error) {
	return s.repo.ListTemplates()
}

func (s *Service) GetTemplate(id string) (Template, error) {
	if err := validID(id, ErrInvalidTemplateID); err != nil {
		return Template{}, err
	}
	return s.repo.GetTemplate(id)
}

func (s *Service) CreateTemplate(input TemplateInput, userID int) (Template, error) {
	input, err := normalizeInput(input)
	if err != nil {
		return Template{}, err
	}
	id := uuid.NewString()
	if err := s.repo.CreateTemplate(id, input, userID); err != nil {
		return Template{}, err
	}
	return s.repo.GetTemplate(id)
}

func (s *Service) UpdateTemplate(id string, input TemplateInput, version int, userID int) (Template, error) {
	if err := validID(id, ErrInvalidTemplateID); err != nil {
		return Template{}, err
	}
	current, err := s.repo.GetTemplate(id)
	if err != nil {
		return Template{}, err
	}
	input.Kind = current.Kind
	input, err = normalizeInput(input)
	if err != nil {
		return Template{}, err
	}
	if err := s.repo.UpdateTemplate(id, input, version, userID); err != nil {
		return Template{}, err
	}
	return s.repo.GetTemplate(id)
}

func (s *Service) DeleteTemplate(id string) error {
	if err := validID(id, ErrInvalidTemplateID); err != nil {
		return err
	}
	return s.repo.DeleteTemplate(id)
}

func (s *Service) ListAssets() ([]Asset, error) {
	return s.repo.ListAssets()
}

func (s *Service) GetAssetData(id string) (AssetData, error) {
	if err := validID(id, ErrInvalidAssetID); err != nil {
		return AssetData{}, err
	}
	return s.repo.GetAssetData(id)
}

func (s *Service) CreateAsset(name string, data []byte, userID int) (Asset, error) {
	if len(data) > MaxAssetSize {
		return Asset{}, ErrAssetTooLarge
	}
	mime := http.DetectContentType(data)
	if !assetTypes[mime] {
		return Asset{}, ErrUnsupportedAsset
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return Asset{}, ErrUnsupportedAsset
	}
	if config.Width <= 0 || config.Height <= 0 || config.Width > MaxAssetDimension || config.Height > MaxAssetDimension {
		return Asset{}, ErrAssetDimensions
	}

	asset := Asset{
		ID:     uuid.NewString(),
		Name:   assetName(name),
		Mime:   mime,
		Width:  config.Width,
		Height: config.Height,
		Size:   len(data),
	}
	if err := s.repo.CreateAsset(asset, data, userID); err != nil {
		return Asset{}, err
	}
	return asset, nil
}

func (s *Service) DeleteAsset(id string) error {
	if err := validID(id, ErrInvalidAssetID); err != nil {
		return err
	}
	return s.repo.DeleteAsset(id)
}

func assetName(name string) string {
	name = strings.TrimSpace(strings.ReplaceAll(name, "\\", "/"))
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	if !utf8.ValidString(name) {
		name = strings.ToValidUTF8(name, "")
	}
	if utf8.RuneCountInString(name) > maxAssetNameLength {
		name = string([]rune(name)[:maxAssetNameLength])
	}
	if name == "" {
		return "kép"
	}
	return name
}
