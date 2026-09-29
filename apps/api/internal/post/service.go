package post

import (
	"speedcubinghungary/internal/types"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetPosts() ([]Post, error) {
	return s.repo.GetAll()
}
func (s *Service) GetLatestPosts() ([]Post, error) {
	return s.repo.GetLatestPosts()
}

// validID: a posts.id UUID oszlop. Egy nem UUID alakú azonosítót a Postgres
// „invalid input syntax for type uuid" hibával utasítana vissza, ami 500-ként
// menne ki — pedig a KLIENS küldött rosszat. Ezért még a repo előtt szűrjük.
func validID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}
	return nil
}

func (s *Service) GetPostByID(id string) (Post, error) {
	if err := validID(id); err != nil {
		return Post{}, err
	}
	return s.repo.GetByID(id)
}
func (s *Service) CreatePost(tag types.PostTag, title, content string, imagePaths []string) error {
	post := Post{
		ID:         uuid.NewString(),
		Tag:        tag,
		Title:      title,
		Content:    content,
		ImagePaths: imagePaths,
	}
	if err := ValidatePost(post); err != nil {
		return err
	}
	return s.repo.Create(post)
}

// DeletePosts törli a posztokat, és visszaadja a hozzájuk tartozó
// képútvonalakat — azokra a fájlokra a törlés után semmi nem hivatkozik.
func (s *Service) DeletePosts(ids []string) ([]string, error) {
	for _, id := range ids {
		if err := validID(id); err != nil {
			return nil, err
		}
	}
	return s.repo.Delete(ids)
}
func (s *Service) UpdatePost(id string, tag types.PostTag, title, content string, imagePaths []string) error {
	post := Post{
		ID:         id,
		Tag:        tag,
		Title:      title,
		Content:    content,
		ImagePaths: imagePaths,
	}
	if err := ValidatePost(post); err != nil {
		return err
	}
	return s.repo.Update(post)
}
