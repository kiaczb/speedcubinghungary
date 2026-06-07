package auth

import (
	"encoding/json"
	"net/http"
	"os"
	"speedcubinghungary/internal/types"

	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service interface {
	LoginCallback(code string) (string, error)
}

type service struct {
	repo       Repository
	httpClient *http.Client
	tokenURL   string
	meURL      string
}

func NewService(repo Repository) Service {
	return &service{
		repo:       repo,
		httpClient: http.DefaultClient,
		tokenURL:   "https://www.worldcubeassociation.org/oauth/token",
		meURL:      "https://www.worldcubeassociation.org/api/v0/me",
	}
}

func toURLValues(data map[string]string) (values map[string][]string) {
	values = map[string][]string{}
	for k, v := range data {
		values[k] = []string{v}
	}
	return
}

func (s *service) LoginCallback(code string) (string, error) {
	// 1️⃣ Code -> WCA Access Token (belső)
	data := map[string]string{
		"client_id":     os.Getenv("WCA_CLIENT_ID"),
		"client_secret": os.Getenv("WCA_CLIENT_SECRET"),
		"grant_type":    "authorization_code",
		"code":          code,
		"redirect_uri":  os.Getenv("WCA_REDIRECT_URI"),
	}

	resp, err := http.PostForm(
		s.tokenURL,
		toURLValues(data),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	// 2️⃣ /me hívás – backend only
	req, _ := http.NewRequest(
		"GET",
		s.meURL,
		nil,
	)
	req.Header.Add("Authorization", "Bearer "+tokenResp.AccessToken)

	client := &http.Client{}
	meResp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer meResp.Body.Close()

	type WCAAvatar struct {
		URL string `json:"url"`
	}

	type WCAUser struct {
		ID    int    `json:"id"`
		WCAID string `json:"wca_id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		// Avatar WCAAvatar `json:"avatar"`
	}
	type wcaMeResponse struct {
		Me WCAUser `json:"me"`
	}

	var rsp wcaMeResponse
	if err := json.NewDecoder(meResp.Body).Decode(&rsp); err != nil {
		return "", err
	}

	wcaUser := rsp.Me

	// 3️⃣ DB check / create
	user, err := s.repo.GetByWCAID(wcaUser.WCAID)
	if err != nil {
		user, err = s.repo.Create(&User{
			WCAID: wcaUser.WCAID,
			Name:  wcaUser.Name,
			Email: wcaUser.Email,
			Roles: []types.Role{types.ROLE_USER},
		})
		if err != nil {
			return "", err
		}
	}

	// 4️⃣ Saját JWT
	claims := UserClaims{
		UserID: user.ID,
		WCAID:  user.WCAID,
		Name:   user.Name,
		// AvatarURL: user.AvatarURL,
		Roles: user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
