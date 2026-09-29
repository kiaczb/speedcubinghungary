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

// stubRepository: csak azt adja vissza, amit a teszt beállít — a lényeg a
// handler hibaleképezése, nem az SQL.
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

// A nem létező poszt 404, nem 500. Korábban a repo `sql.ErrNoRows`-a végigment
// a handlerig, és „sql: no rows in result set" szöveggel adott szerverhibát —
// a kliens nem tudta megkülönböztetni a rossz azonosítót a leállt adatbázistól.
func TestGetPostByID_MissingPostIs404(t *testing.T) {
	recorder := getPostByID(t, &stubRepository{err: ErrPostNotFound}, knownID)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

// A posts.id UUID: egy nem UUID alakú azonosítóra a Postgres „invalid input
// syntax for type uuid" hibát ad. Ez a KLIENS hibája (400), és a szövege sem
// mehet ki — az oszlop típusát árulná el.
func TestGetPostByID_MalformedIDIs400(t *testing.T) {
	recorder := getPostByID(t, &stubRepository{}, "nem-uuid")

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ErrInvalidID.Error())
}

// Valódi adatbázishiba: 500, de a szövege a naplóé, nem a böngészőé.
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

// A DELETE törzse kliens-adat: egy nem UUID azonosító nem juthat le a
// `WHERE id = ANY($1)`-ig, mert ott nyers típushibát okozna.
func TestDeletePosts_MalformedIDIs400(t *testing.T) {
	handler := NewHandler(NewService(&stubRepository{}), t.TempDir())

	recorder := httptest.NewRecorder()
	handler.DeletePosts(recorder, httptest.NewRequest(
		http.MethodDelete, "/posts", strings.NewReader(`{"ids":["nem-uuid"]}`)))

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// A repository szerződése: a hiányzó sorból ErrPostNotFound lesz, nem
// sql.ErrNoRows — erre épül a handler 404-e.
func TestErrPostNotFound_IsDistinctFromSQLNoRows(t *testing.T) {
	assert.False(t, errors.Is(ErrPostNotFound, sql.ErrNoRows))
}

// ⚠️ A poszt törlése korábban csak a DB sort vitte el, a képfájlokat nem: a
// galéria óta ez posztonként akár 10 fájl, amire soha többé semmi nem
// hivatkozik, és amit csak kézzel lehetett kitakarítani a volume-ról.
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

// Ha a törlés félúton hibázik, a DB sorok egy része már eltűnt — a
// visszakapott útvonalak fájljait akkor is el kell takarítani.
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
