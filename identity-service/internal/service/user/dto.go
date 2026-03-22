package user

type ChangePasswordInput struct {
	OldPassword string
	NewPassword string
}

type LinkIdentityInput struct {
	Provider    string
	Code        string
	RedirectURI string
}
