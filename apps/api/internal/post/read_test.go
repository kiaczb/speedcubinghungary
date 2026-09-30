package post

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"speedcubinghungary/internal/types"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRepository struct {
	post    Post
	deleted []string
	err     error
}

func (s *stubRepository) GetAll() ([]Post, error)           { return nil, s.err }
func (s *stubRepository) GetLatestPosts() ([]Post, error)   { return nil, s.err }
func (s *stubRepository) GetByID(string) (Post, error)      { return s.post, s.err }
func (s *stubRepository) Create(Post) error                 { return s.err }
func (s *stubRepository) Update(Post) error                 { return s.err }
func (s *stubRepository) Delete([]string) ([]string, error) { return s.deleted, s.err }

func getPostByID(t *testing.T, repo Repository, id string) *httptest.ResponseRecorder {
	t.Helper()

	router := chi.NewRouter()
	router.Get("/posts/{id}", NewHandler(NewService(repo), t.TempDir()).GetPostByID)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/posts/"+id, nil))

	return recorder
}

const knownID = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"

func TestGetPostByID_MissingPostIs404(t *testing.T) {
	recorder := getPostByID(t, &stubRepository{err: ErrPostNotFound}, knownID)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestGetPostByID_MalformedIDIs400(t *testing.T) {
	recorder := getPostByID(t, &stubRepository{}, "nem-uuid")

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ErrInvalidID.Error())
}

func TestGetPostByID_DatabaseErrorDoesNotLeak(t *testing.T) {
	leaky := errors.New(`pq: relation "posts" does not exist`)
	recorder := getPostByID(t, &stubRepository{err: leaky}, knownID)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "relation")
}

func TestGetPostByID_ServesTheFoundPost(t *testing.T) {
	repo := &stubRepository{post: Post{ID: knownID, Tag: types.TAG_NEWS, Title: "Teszt"}}

	recorder := getPostByID(t, repo, knownID)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), knownID)
}

func TestDeletePosts_MalformedIDIs400(t *testing.T) {
	handler := NewHandler(NewService(&stubRepository{}), t.TempDir())

	recorder := httptest.NewRecorder()
	handler.DeletePosts(recorder, httptest.NewRequest(
		http.MethodDelete, "/posts", strings.NewReader(`{"ids":["nem-uuid"]}`)))

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestErrPostNotFound_IsDistinctFromSQLNoRows(t *testing.T) {
	assert.False(t, errors.Is(ErrPostNotFound, sql.ErrNoRows))
}

func TestDeletePosts_RemovesTheImageFiles(t *testing.T) {
	uploadDir := t.TempDir()

	kept := writeUpload(t, uploadDir, "masik-poszte.png")
	removed := []string{
		writeUpload(t, uploadDir, "borito.png"),
		writeUpload(t, uploadDir, "masodik.png"),
	}

	handler := NewHandler(NewService(&stubRepository{deleted: removed}), uploadDir)

	recorder := httptest.NewRecorder()
	handler.DeletePosts(recorder, httptest.NewRequest(
		http.MethodDelete, "/posts", strings.NewReader(`{"ids":["`+knownID+`"]}`)))

	require.Equal(t, http.StatusOK, recorder.Code)

	for _, path := range removed {
		assert.NoFileExists(t, filepath.Join(uploadDir, filepath.Base(path)),
			"a törölt poszt képére már semmi nem hivatkozik")
	}
	assert.FileExists(t, filepath.Join(uploadDir, filepath.Base(kept)),
		"másik poszt képéhez nem nyúlunk")
}

func TestDeletePosts_CleansUpEvenWhenTheDeleteFails(t *testing.T) {
	uploadDir := t.TempDir()
	orphan := writeUpload(t, uploadDir, "felig-torolt.png")

	repo := &stubRepository{deleted: []string{orphan}, err: errors.New("pq: connection reset")}
	handler := NewHandler(NewService(repo), uploadDir)

	recorder := httptest.NewRecorder()
	handler.DeletePosts(recorder, httptest.NewRequest(
		http.MethodDelete, "/posts", strings.NewReader(`{"ids":["`+knownID+`"]}`)))

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.NoFileExists(t, filepath.Join(uploadDir, filepath.Base(orphan)))
}

func writeUpload(t *testing.T, uploadDir, name string) string {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(uploadDir, name), []byte("kep"), 0o644))
	return PublicUploadPath + "/" + name
}
