package post

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const (
	// MaxImageSize egy feltöltött kép maximális mérete.
	MaxImageSize = 5 << 20 // 5 MiB

	// MaxImages a poszthoz csatolható képek maximális száma. A frontend
	// ugyanezt a korlátot mutatja (MAX_POST_IMAGES) — ha itt változik, ott is.
	MaxImages = 10

	// PublicUploadPath az az útvonal, amin az API kiszolgálja a képeket; az
	// image_paths elemei ezzel a prefixszel kerülnek a DB-be.
	PublicUploadPath = "/uploads"
)

var (
	ErrUnsupportedImage = errors.New("unsupported image type (allowed: jpeg, png, webp, gif)")
	ErrTooManyImages    = fmt.Errorf("a post can have at most %d images", MaxImages)
	ErrImageTooLarge    = fmt.Errorf("image is too large (max %d MiB)", MaxImageSize>>20)
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

func SaveImage(file multipart.File, uploadDir string) (string, error) {
	head := make([]byte, 512)
	n, err := file.Read(head)
	if err != nil && err != io.EOF {
		return "", err
	}

	ext, ok := allowedImageTypes[http.DetectContentType(head[:n])]
	if !ok {
		return "", ErrUnsupportedImage
	}

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

func SaveImages(headers []*multipart.FileHeader, uploadDir string) ([]string, error) {
	if len(headers) > MaxImages {
		return nil, ErrTooManyImages
	}

	paths := make([]string, 0, len(headers))
	fail := func(err error) ([]string, error) {
		RemoveImages(paths, uploadDir)
		return nil, err
	}

	for _, header := range headers {
		// A méretkorlátot fájlonként a FileHeader alapján utasítjuk el: a
		// SaveImage-beli LimitReader csak CSONKÍTANA, a túl nagy kép pedig
		// legyen 400-as hiba, ne néma félkép.
		if header.Size > MaxImageSize {
			return fail(ErrImageTooLarge)
		}

		file, err := header.Open()
		if err != nil {
			return fail(err)
		}

		path, err := SaveImage(file, uploadDir)
		file.Close()
		if err != nil {
			return fail(err)
		}

		paths = append(paths, path)
	}

	return paths, nil
}

// RemoveImage törli a publikus útvonalhoz tartozó fájlt. A filepath.Base miatt
// csak az uploadDir-en belül törölhet.
func RemoveImage(publicPath, uploadDir string) {
	if publicPath == "" {
		return
	}
	os.Remove(filepath.Join(uploadDir, filepath.Base(publicPath)))
}

func RemoveImages(publicPaths []string, uploadDir string) {
	for _, path := range publicPaths {
		RemoveImage(path, uploadDir)
	}
}
