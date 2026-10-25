package auth

import "preditto/internal/model"

type TokenResponse struct {
	// Access JWT for Authorization: Bearer <access_token>.
	AccessToken string `json:"access_token"`
	// Refresh JWT for the refresh/logout request body only.
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type" example:"Bearer"`
	// Access token lifetime in seconds.
	ExpiresIn int64 `json:"expires_in" format:"int64" example:"900"`
}

type AuthResponse struct {
	TokenResponse
	User model.User `json:"user"`
}
