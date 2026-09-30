package post

import (
	"testing"

	"speedcubinghungary/internal/types"

	"github.com/stretchr/testify/assert"
)

func TestValidatePost_Tag(t *testing.T) {
	const validContent = "elég hosszú tartalom a validációhoz"

	tests := []struct {
		name    string
		tag     types.PostTag
		wantErr error
	}{
		{"event", types.TAG_EVENT, nil},
		{"competition", types.TAG_COMPETITION, nil},
		{"news", types.TAG_NEWS, nil},
		{"announcement", types.TAG_ANNOUNCEMENT, nil},
		{"report", types.TAG_REPORT, nil},
		{"empty", types.PostTag(""), ErrInvalidTag},
		{"unknown", types.PostTag("blogpost"), ErrInvalidTag},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePost(Post{
				ID:      "5f9d2a1e-0000-0000-0000-000000000000",
				Tag:     tt.tag,
				Title:   "Cím",
				Content: validContent,
			})
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestValidatePost_RequiredFields(t *testing.T) {
	base := Post{
		ID:      "5f9d2a1e-0000-0000-0000-000000000000",
		Tag:     types.TAG_REPORT,
		Title:   "Heti riport",
		Content: "elég hosszú tartalom a validációhoz",
	}

	t.Run("missing id", func(t *testing.T) {
		p := base
		p.ID = ""
		assert.Equal(t, ErrInvalidID, ValidatePost(p))
	})

	t.Run("empty title", func(t *testing.T) {
		p := base
		p.Title = ""
		assert.Equal(t, ErrEmptyTitle, ValidatePost(p))
	})

	t.Run("content shorter than 20 chars", func(t *testing.T) {
		p := base
		p.Content = "rövid"
		assert.Equal(t, ErrShortContent, ValidatePost(p))
	})

	t.Run("valid post passes", func(t *testing.T) {
		assert.NoError(t, ValidatePost(base))
	})
}
