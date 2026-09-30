package post

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveImageOrder(t *testing.T) {
	existing := []string{"/uploads/a.jpg", "/uploads/b.jpg"}
	uploaded := []string{"/uploads/c.jpg", "/uploads/d.jpg"}

	tests := []struct {
		name     string
		raw      string
		existing []string
		uploaded []string
		want     []string
		wantErr  error
	}{
		{
			name:     "megtartás sorrendben",
			raw:      `["/uploads/a.jpg","/uploads/b.jpg"]`,
			existing: existing,
			want:     []string{"/uploads/a.jpg", "/uploads/b.jpg"},
		},
		{
			name:     "átrendezés",
			raw:      `["/uploads/b.jpg","/uploads/a.jpg"]`,
			existing: existing,
			want:     []string{"/uploads/b.jpg", "/uploads/a.jpg"},
		},
		{
			name:     "törlés",
			raw:      `["/uploads/b.jpg"]`,
			existing: existing,
			want:     []string{"/uploads/b.jpg"},
		},
		{
			name:     "minden kép törlése",
			raw:      `[]`,
			existing: existing,
			want:     []string{},
		},
		{
			name:     "új kép beszúrása a meglévők közé",
			raw:      `["/uploads/a.jpg","new","/uploads/b.jpg","new"]`,
			existing: existing,
			uploaded: uploaded,
			want: []string{
				"/uploads/a.jpg", "/uploads/c.jpg",
				"/uploads/b.jpg", "/uploads/d.jpg",
			},
		},
		{
			name:     "idegen útvonal elutasítva",
			raw:      `["/uploads/masik-poszt.jpg"]`,
			existing: existing,
			wantErr:  ErrInvalidImageOrder,
		},
		{
			name:     "duplikált útvonal elutasítva",
			raw:      `["/uploads/a.jpg","/uploads/a.jpg"]`,
			existing: existing,
			wantErr:  ErrInvalidImageOrder,
		},
		{
			name:     "hivatkozás nélküli feltöltés elutasítva",
			raw:      `["/uploads/a.jpg"]`,
			existing: existing,
			uploaded: uploaded,
			wantErr:  ErrInvalidImageOrder,
		},
		{
			name:     "több jelölő, mint feltöltött kép",
			raw:      `["new","new"]`,
			existing: existing,
			uploaded: uploaded[:1],
			wantErr:  ErrInvalidImageOrder,
		},
		{
			name:     "érvénytelen json",
			raw:      `nem json`,
			existing: existing,
			wantErr:  ErrInvalidImageOrder,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveImageOrder(tt.raw, tt.existing, tt.uploaded)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveImageOrder_MaxImages(t *testing.T) {
	existing := make([]string, 0, MaxImages+1)
	raw := "["
	for i := 0; i <= MaxImages; i++ {
		path := "/uploads/" + string(rune('a'+i)) + ".jpg"
		existing = append(existing, path)
		if i > 0 {
			raw += ","
		}
		raw += `"` + path + `"`
	}
	raw += "]"

	got, err := resolveImageOrder(raw, existing, nil)
	assert.ErrorIs(t, err, ErrTooManyImages)
	assert.Nil(t, got)
}
