package certificate

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"strconv"
	"testing"

	"speedcubinghungary/internal/types"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmokeDatabase_TemplatesAndAssets(t *testing.T) {
	url := os.Getenv("SMOKE_DATABASE_URL")
	if url == "" {
		t.Skip("SMOKE_DATABASE_URL nincs megadva")
	}
	userID, err := strconv.Atoi(os.Getenv("SMOKE_USER_ID"))
	require.NoError(t, err, "SMOKE_USER_ID=<users.id> kell")

	db, err := sql.Open("pgx", url)
	require.NoError(t, err)
	defer db.Close()
	service := NewService(NewPostgresRepository(db), nil, nil)

	var pixel bytes.Buffer
	require.NoError(t, png.Encode(&pixel, image.NewRGBA(image.Rect(0, 0, 3, 2))))
	asset, err := service.CreateAsset("C:\\képek\\logó.png", pixel.Bytes(), userID)
	require.NoError(t, err)
	assert.Equal(t, "logó.png", asset.Name)
	assert.Equal(t, [2]int{3, 2}, [2]int{asset.Width, asset.Height})
	defer service.DeleteAsset(asset.ID)

	data, err := service.GetAssetData(asset.ID)
	require.NoError(t, err)
	assert.Equal(t, "image/png", data.Mime)
	assert.Equal(t, pixel.Bytes(), data.Data)

	document := func(assetID string) json.RawMessage {
		return json.RawMessage(`{"schemaVersion":1,"page":{"width":297,"height":210},"elements":[{"id":"a","type":"image","assetId":"` + assetID + `"}]}`)
	}
	created, err := service.CreateTemplate(TemplateInput{
		Name: "  Füstteszt  ", Kind: types.CERTIFICATE_PLACEMENT, Document: document(asset.ID),
	}, userID)
	require.NoError(t, err)
	defer service.DeleteTemplate(created.ID)
	assert.Equal(t, "Füstteszt", created.Name)
	assert.Equal(t, 1, created.Version)
	assert.NotEmpty(t, created.UpdatedByName)

	assert.ErrorIs(t, service.DeleteAsset(asset.ID), ErrAssetInUse, "sablonban használt kép nem törölhető")

	updated, err := service.UpdateTemplate(created.ID, TemplateInput{
		Name: "Átnevezve", Kind: types.CERTIFICATE_PARTICIPATION, Document: document("none"),
	}, created.Version, userID)
	require.NoError(t, err)
	assert.Equal(t, 2, updated.Version)
	assert.Equal(t, types.CERTIFICATE_PLACEMENT, updated.Kind, "a típus létrehozás után nem változik")

	_, err = service.UpdateTemplate(created.ID, TemplateInput{Name: "Régi", Document: document("none")}, created.Version, userID)
	assert.ErrorIs(t, err, ErrVersionConflict)

	require.NoError(t, service.DeleteAsset(asset.ID), "már egy sablon sem használja")
	require.NoError(t, service.DeleteTemplate(created.ID))
	assert.ErrorIs(t, service.DeleteTemplate(created.ID), ErrTemplateNotFound)
	_, err = service.UpdateTemplate(created.ID, TemplateInput{Name: "x", Document: document("none")}, 2, userID)
	assert.ErrorIs(t, err, ErrTemplateNotFound)
}
