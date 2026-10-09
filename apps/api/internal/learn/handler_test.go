package learn

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitPath(t *testing.T) {
	assert.Equal(t, []string{}, splitPath(""))
	assert.Equal(t, []string{}, splitPath("/"))
	assert.Equal(t, []string{"3x3", "cfop"}, splitPath("/3x3/cfop/"))
	assert.Equal(t, []string{"3x3", "", "cfop"}, splitPath("3x3//cfop"))
}

func TestGetLevel_Statuses(t *testing.T) {
	repo := &stubRepository{chain: []Node{{ID: "n1", Slug: "3x3"}}, children: []Node{}}
	handler := NewHandler(NewService(repo))

	cases := map[string]int{
		"/learn/level":                        http.StatusOK,
		"/learn/level?path=3x3":               http.StatusOK,
		"/learn/level?path=3x3%2F%2Fcfop":     http.StatusBadRequest,
		"/learn/level?path=3x3%2Fnincs-ilyen": http.StatusNotFound,
	}

	for url, expected := range cases {
		recorder := httptest.NewRecorder()
		handler.GetLevel(recorder, httptest.NewRequest(http.MethodGet, url, nil))
		assert.Equal(t, expected, recorder.Code, url)
	}
}

func TestGetLevel_EncodesEmptyListsNotNull(t *testing.T) {
	repo := &stubRepository{chain: []Node{}, children: []Node{}}
	recorder := httptest.NewRecorder()

	NewHandler(NewService(repo)).GetLevel(recorder, httptest.NewRequest(http.MethodGet, "/learn/level", nil))

	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	assert.Equal(t, "null", string(body["node"]))
	assert.Equal(t, "[]", string(body["ancestors"]))
	assert.Equal(t, "[]", string(body["children"]))
	assert.Equal(t, "[]", string(body["algorithms"]))
}

func TestErrorStatus(t *testing.T) {
	cases := map[error]int{
		ErrEmptyTitle:               http.StatusBadRequest,
		ErrInvalidPath:              http.StatusBadRequest,
		ErrNodeNotFound:             http.StatusNotFound,
		ErrParentNotFound:           http.StatusNotFound,
		ErrAlgorithmNotFound:        http.StatusNotFound,
		ErrSlugTaken:                http.StatusConflict,
		ErrParentHasAlgorithms:      http.StatusConflict,
		ErrNodeHasChildren:          http.StatusConflict,
		ErrMoveIntoOwnSubtree:       http.StatusConflict,
		ErrTooDeep:                  http.StatusConflict,
		ErrOrderMismatch:            http.StatusConflict,
		errors.New("pq: kapcsolat"): http.StatusInternalServerError,
	}

	for err, expected := range cases {
		assert.Equal(t, expected, errorStatus(err), err.Error())
	}
}
