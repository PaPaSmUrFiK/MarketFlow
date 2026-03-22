package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"io"
	"net/http"
)

type UserInfo struct {
	ID    string // ID пользователя на стороне провайдера — используем как ProviderUserID
	Email string // может быть пустым если провайдер не отдал
	Name  string
}

type googleProvider struct {
	config *oauth2.Config
}

type githubProvider struct {
	config *oauth2.Config
}

type Provider interface {
	// AuthURL возвращает URL для редиректа пользователя на страницу логина провайдера.
	// state — случайная строка для CSRF защиты, генерируется на стороне клиента.
	AuthURL(state string) string

	// GetUserInfo обменивает одноразовый code на профиль пользователя.
	// code живёт секунды — нельзя повторно использовать.
	GetUserInfo(ctx context.Context, code string) (*UserInfo, error)
}

func NewGoogleProvider(clientID, clientSecret, redirectURI string) Provider {
	return &googleProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURI,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (p *googleProvider) AuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (p *googleProvider) GetUserInfo(ctx context.Context, code string) (*UserInfo, error) {
	const op = "oauth.googleProvider.GetUserInfo"

	// Обмениваем code → access token
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%s: exchange code: %w", op, err)
	}

	// Получаем профиль через Google userinfo endpoint
	client := p.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("%s: get userinfo: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: userinfo status %d", op, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: read body: %w", op, err)
	}

	var info struct {
		Sub   string `json:"sub"` // Google user ID — стабильный, не меняется
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("%s: decode response: %w", op, err)
	}

	if info.Sub == "" {
		return nil, fmt.Errorf("%s: empty user id from google", op)
	}

	return &UserInfo{
		ID:    info.Sub,
		Email: info.Email,
		Name:  info.Name,
	}, nil
}

func NewGitHubProvider(clientID, clientSecret, redirectURI string) Provider {
	return &githubProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURI,
			Scopes:       []string{"user:email", "read:user"},
			Endpoint:     github.Endpoint,
		},
	}
}

func (p *githubProvider) AuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (p *githubProvider) GetUserInfo(ctx context.Context, code string) (*UserInfo, error) {
	const op = "oauth.githubProvider.GetUserInfo"

	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%s: exchange code: %w", op, err)
	}

	client := p.config.Client(ctx, token)

	// Получаем базовый профиль
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("%s: get user: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: user status %d", op, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: read body: %w", op, err)
	}

	var info struct {
		ID    int64  `json:"id"`    // GitHub user ID — числовой, стабильный
		Email string `json:"email"` // может быть пустым если приватный
		Name  string `json:"name"`
		Login string `json:"login"` // username
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("%s: decode response: %w", op, err)
	}

	if info.ID == 0 {
		return nil, fmt.Errorf("%s: empty user id from github", op)
	}

	// Если email приватный — пробуем получить через отдельный endpoint
	email := info.Email
	if email == "" {
		email, err = p.getPrimaryEmail(ctx, client)
		if err != nil {
			// Не фатально — продолжаем без email
			email = ""
		}
	}

	return &UserInfo{
		ID:    fmt.Sprintf("%d", info.ID), // конвертируем int64 → string как у Google
		Email: email,
		Name:  info.Name,
	}, nil
}

// getPrimaryEmail получает основной email пользователя GitHub.
// Нужен когда email скрыт в основном профиле.
func (p *githubProvider) getPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	return "", nil
}
