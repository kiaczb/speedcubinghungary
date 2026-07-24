package post

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const (
	// MaxImageSize a feltölthető borítókép maximális mérete.
	MaxImageSize = 5 << 20 // 5 MiB

	// PublicUploadPath az az útvonal, amin az API kiszolgálja a képeket.
	// A DB-be ezzel a prefixszel mentjük az image_path értéket.
	PublicUploadPath = "/uploads"
)

var (
	ErrUnsupportedImage = errors.New("unsupported image type (allowed: jpeg, png, webp, gif)")
)

// allowedImageTypes a ténylegesen elfogadott MIME típusok és a hozzájuk tartozó
// kiterjesztés. A típust a fájl TARTALMÁBÓL detektáljuk (http.DetectContentType),
// nem a kliens által küldött Content-Type fejlécből — azt bárki hazudhatja.
var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// SaveImage kiírja a feltöltött képet az uploadDir-be egy generált néven, és
// visszaadja a publikus útvonalát (pl. "/uploads/<uuid>.webp").
// A kliens által küldött fájlnevet szándékosan eldobjuk: saját uuid nevet adunk,
// így nem lehet path traversallal (../) kitörni az uploadDir-ből.
func SaveImage(file multipart.File, uploadDir string) (string, error) {
	// Az első 512 byte elég a típus detektálásához.
	head := make([]byte, 512)
	n, err := file.Read(head)
	if err != nil && err != io.EOF {
		return "", err
	}

	ext, ok := allowedImageTypes[http.DetectContentType(head[:n])]
	if !ok {
		return "", ErrUnsupportedImage
	}

	// Vissza a fájl elejére, hogy a teljes tartalmat ki tudjuk írni.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return "", err
	}

	name := uuid.NewString() + ext
	dst, err := os.Create(filepath.Join(uploadDir, name))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// LimitReader: a méretkorlátot itt is betartatjuk, ne csak a kérés szintjén.
	if _, err := io.Copy(dst, io.LimitReader(file, MaxImageSize)); err != nil {
		return "", err
	}

	return PublicUploadPath + "/" + name, nil
}

// RemoveImage törli a publikus útvonalhoz tartozó fájlt. Akkor használjuk, ha a
// kép már kiírásra került, de a poszt mentése elbukott — így nem marad árva
// fájl a volume-on. filepath.Base miatt csak az uploadDir-en belül töröl.
func RemoveImage(publicPath, uploadDir string) {
	if publicPath == "" {
		return
	}
	os.Remove(filepath.Join(uploadDir, filepath.Base(publicPath)))
}
