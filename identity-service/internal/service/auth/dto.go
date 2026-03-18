package auth

// RegisterInput содержит данные для регистрации пользователя.
type RegisterInput struct {
	AppCode   string
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

// LoginInput содержит данные для аутентификации пользователя.
type LoginInput struct {
	AppCode   string
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

// OAuthLoginInput содержит данные для OAuth-аутентификации.
type OAuthLoginInput struct {
	AppCode      string
	Provider     string
	Code         string
	RedirectURI  string
	UserAgent    string
	IPAddress    string
}

// TokenPair представляет пару токенов (access + refresh).
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}
