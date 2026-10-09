package learn

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"OLL 24":                 "oll-24",
		"3x3":                    "3x3",
		"  Skewb / NS  ":         "skewb-ns",
		"Ősz – Különleges eset":  "osz-kulonleges-eset",
		"R U R' U'":              "r-u-r-u",
		"Árvíztűrő tükörfúrógép": "arvizturo-tukorfurogep",
		"!!!":                    "",
		"Кубик":                  "",
	}

	for input, expected := range cases {
		assert.Equal(t, expected, Slugify(input), input)
	}
}

func TestSlugify_TruncatesWithoutATrailingHyphen(t *testing.T) {
	slug := Slugify(strings.Repeat("a", MaxSlugLength-1) + " bcd")

	assert.LessOrEqual(t, len(slug), MaxSlugLength)
	assert.False(t, strings.HasSuffix(slug, "-"))
}

func TestNormalizeNode_DerivesTheSlugFromTheTitle(t *testing.T) {
	in, err := normalizeNode(NodeInput{Title: "  OLL 24 ", Description: " Kamaleon "})

	require.NoError(t, err)
	assert.Equal(t, "OLL 24", in.Title)
	assert.Equal(t, "oll-24", in.Slug)
	assert.Equal(t, "Kamaleon", in.Description)
}

func TestNormalizeNode_NormalizesAGivenSlug(t *testing.T) {
	in, err := normalizeNode(NodeInput{Title: "Harmadik kocka", Slug: "Három X Három"})

	require.NoError(t, err)
	assert.Equal(t, "harom-x-harom", in.Slug)
}

func TestNormalizeNode_CanonicalizesTheParentID(t *testing.T) {
	parent := "6F9619FF-8B86-D011-B42D-00C04FC964FF"

	in, err := normalizeNode(NodeInput{ParentID: &parent, Title: "CFOP"})

	require.NoError(t, err)
	assert.Equal(t, "6f9619ff-8b86-d011-b42d-00c04fc964ff", *in.ParentID)
}

func TestNormalizeNode_Rejections(t *testing.T) {
	notUUID := "nem-uuid"
	cases := map[string]struct {
		in       NodeInput
		expected error
	}{
		"üres cím":           {NodeInput{Title: "   "}, ErrEmptyTitle},
		"túl hosszú cím":     {NodeInput{Title: strings.Repeat("x", MaxTitleLength+1)}, ErrTitleTooLong},
		"slug nem képezhető": {NodeInput{Title: "!!!"}, ErrInvalidSlug},
		"hibás szülő":        {NodeInput{ParentID: &notUUID, Title: "CFOP"}, ErrInvalidParentID},
		"túl hosszú leírás":  {NodeInput{Title: "CFOP", Description: strings.Repeat("x", MaxDescriptionLength+1)}, ErrDescriptionTooLong},
	}

	for name, c := range cases {
		_, err := normalizeNode(c.in)
		assert.ErrorIs(t, err, c.expected, name)
		assert.ErrorIs(t, err, ErrValidation, name)
	}
}

func TestNormalizeNode_Icon(t *testing.T) {
	for _, icon := range []string{"", "event-333", " event-skewb ", "unofficial-fto", "unofficial-333_oven_mitts"} {
		_, err := normalizeNode(NodeInput{Title: "3x3", Icon: icon})
		assert.NoError(t, err, icon)
	}

	for _, icon := range []string{"penalty-A3d1", "333", "event-", "event-333 evil", "event-333\"><script>", "EVENT-333"} {
		_, err := normalizeNode(NodeInput{Title: "3x3", Icon: icon})
		assert.ErrorIs(t, err, ErrInvalidIcon, icon)
	}
}

func TestNormalizeNode_ImageURL(t *testing.T) {
	for _, image := range []string{
		"",
		" https://visualcube.api.cubing.net/visualcube.php?fmt=svg&pzl=3&stage=oll&case=r%20U%20R%27 ",
		"https://example.com/oll-24.png",
	} {
		_, err := normalizeNode(NodeInput{Title: "OLL 24", ImageURL: image})
		assert.NoError(t, err, image)
	}

	for _, image := range []string{
		"http://example.com/oll.png",
		"javascript:alert(1)",
		"data:image/svg+xml;base64,PHN2Zz4=",
		"/uploads/kep.png",
		"https://",
		"https://user:pass@example.com/oll.png",
		"https://example.com/a b.png",
		"https://example.com/\"onerror=\"x",
		"https://example.com/" + strings.Repeat("a", MaxImageURLLength),
	} {
		_, err := normalizeNode(NodeInput{Title: "OLL 24", ImageURL: image})
		assert.ErrorIs(t, err, ErrInvalidImageURL, image)
	}
}

func TestNormalizeAlgorithm_CollapsesWhitespace(t *testing.T) {
	in, err := normalizeAlgorithm(AlgorithmInput{Notation: "  r U  R'\n U'\tr' F R F' ", Name: " Fő "})

	require.NoError(t, err)
	assert.Equal(t, "r U R' U' r' F R F'", in.Notation)
	assert.Equal(t, "Fő", in.Name)
}

func TestNormalizeAlgorithm_RejectsAnEmptyNotation(t *testing.T) {
	_, err := normalizeAlgorithm(AlgorithmInput{Notation: " \n "})

	assert.ErrorIs(t, err, ErrEmptyNotation)
}

func TestValidatePath(t *testing.T) {
	assert.NoError(t, validatePath([]string{}))
	assert.NoError(t, validatePath([]string{"3x3", "cfop", "oll", "oll-24"}))

	tooDeep := make([]string, MaxDepth+1)
	for i := range tooDeep {
		tooDeep[i] = "a"
	}

	for _, path := range [][]string{
		{"3x3", ""},
		{".."},
		{"OLL"},
		{"oll--24"},
		{strings.Repeat("a", MaxSlugLength+1)},
		tooDeep,
	} {
		assert.ErrorIs(t, validatePath(path), ErrInvalidPath, "%v", path)
	}
}

func TestNormalizeOrder(t *testing.T) {
	a := "6F9619FF-8B86-D011-B42D-00C04FC964FF"
	b := "7c9e6679-7425-40de-944b-e07fc1f90ae7"

	ids, err := normalizeOrder([]string{a, b})
	require.NoError(t, err)
	assert.Equal(t, []string{"6f9619ff-8b86-d011-b42d-00c04fc964ff", b}, ids)

	_, err = normalizeOrder([]string{b, b})
	assert.ErrorIs(t, err, ErrInvalidOrder, "ugyanaz az id kétszer")

	_, err = normalizeOrder([]string{"nem-uuid"})
	assert.ErrorIs(t, err, ErrInvalidOrder)

	_, err = normalizeOrder([]string{"{" + b + "}"})
	assert.ErrorIs(t, err, ErrInvalidOrder, "csak a kanonikus alak")
}
