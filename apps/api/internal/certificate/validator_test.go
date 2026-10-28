package certificate

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"strings"
	"testing"

	"speedcubinghungary/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validDocumentJSON = `{"schemaVersion":1,"page":{"width":297,"height":210},"elements":[]}`

func TestNormalizeInput(t *testing.T) {
	valid := TemplateInput{Name: " Dobogós ", Kind: types.CERTIFICATE_PLACEMENT, Document: json.RawMessage("{\n \"schemaVersion\": 1, \"page\": {}, \"elements\": []\n}")}
	normalized, err := normalizeInput(valid)
	require.NoError(t, err)
	assert.Equal(t, "Dobogós", normalized.Name)
	assert.Equal(t, `{"schemaVersion":1,"page":{},"elements":[]}`, string(normalized.Document))

	cases := map[string]struct {
		input TemplateInput
		err   error
	}{
		"üres név":         {TemplateInput{Name: "  ", Kind: types.CERTIFICATE_PLACEMENT, Document: json.RawMessage(validDocumentJSON)}, ErrEmptyName},
		"túl hosszú név":   {TemplateInput{Name: strings.Repeat("é", MaxNameLength+1), Kind: types.CERTIFICATE_PLACEMENT, Document: json.RawMessage(validDocumentJSON)}, ErrNameTooLong},
		"ismeretlen típus": {TemplateInput{Name: "x", Kind: "diploma", Document: json.RawMessage(validDocumentJSON)}, ErrInvalidKind},
		"nem objektum":     {TemplateInput{Name: "x", Kind: types.CERTIFICATE_PLACEMENT, Document: json.RawMessage(`[]`)}, ErrInvalidDocument},
		"rossz séma":       {TemplateInput{Name: "x", Kind: types.CERTIFICATE_PLACEMENT, Document: json.RawMessage(`{"schemaVersion":2,"page":{},"elements":[]}`)}, ErrInvalidDocument},
		"hiányzó elemek":   {TemplateInput{Name: "x", Kind: types.CERTIFICATE_PLACEMENT, Document: json.RawMessage(`{"schemaVersion":1,"page":{}}`)}, ErrInvalidDocument},
		"utána szemét":     {TemplateInput{Name: "x", Kind: types.CERTIFICATE_PLACEMENT, Document: json.RawMessage(validDocumentJSON + `{}`)}, ErrInvalidDocument},
		"túl sok elem": {TemplateInput{Name: "x", Kind: types.CERTIFICATE_PLACEMENT,
			Document: json.RawMessage(`{"schemaVersion":1,"page":{},"elements":[` + strings.TrimSuffix(strings.Repeat(`{},`, MaxElements+1), ",") + `]}`)}, ErrTooManyElements},
	}
	for name, c := range cases {
		_, err := normalizeInput(c.input)
		assert.ErrorIs(t, err, c.err, name)
		assert.ErrorIs(t, err, ErrValidation, name)
	}
}

type recordingRepository struct {
	Repository
	assets []Asset
}

func (r *recordingRepository) CreateAsset(asset Asset, _ []byte, _ int) error {
	r.assets = append(r.assets, asset)
	return nil
}

func TestCreateAsset(t *testing.T) {
	repo := &recordingRepository{}
	service := NewService(repo, nil, nil)

	var pngBytes bytes.Buffer
	require.NoError(t, png.Encode(&pngBytes, image.NewRGBA(image.Rect(0, 0, 4, 3))))
	asset, err := service.CreateAsset("../../etc/passwd", pngBytes.Bytes(), 1)
	require.NoError(t, err)
	assert.Equal(t, "passwd", asset.Name, "a kliens útvonala nem kerül a névbe")
	assert.Equal(t, "image/png", asset.Mime)
	assert.Equal(t, 4, asset.Width)

	var gifBytes bytes.Buffer
	require.NoError(t, gif.Encode(&gifBytes, image.NewPaletted(image.Rect(0, 0, 1, 1), color.Palette{color.Black}), nil))
	_, err = service.CreateAsset("a.gif", gifBytes.Bytes(), 1)
	assert.ErrorIs(t, err, ErrUnsupportedAsset)

	_, err = service.CreateAsset("a.png", []byte("<svg xmlns='http://www.w3.org/2000/svg'/>"), 1)
	assert.ErrorIs(t, err, ErrUnsupportedAsset)

	_, err = service.CreateAsset("a.png", make([]byte, MaxAssetSize+1), 1)
	assert.ErrorIs(t, err, ErrAssetTooLarge)

	assert.Len(t, repo.assets, 1)
}

func TestTemplateIDsAreValidatedBeforeTheDatabase(t *testing.T) {
	service := NewService(nil, nil, nil)
	_, err := service.GetTemplate("1; DROP TABLE x")
	assert.ErrorIs(t, err, ErrInvalidTemplateID)
	assert.ErrorIs(t, service.DeleteAsset("nem-uuid"), ErrInvalidAssetID)
}
