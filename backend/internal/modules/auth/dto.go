package auth

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string
	SessionID    string
	PublicID     string
}

type LogoutRequest struct {
	RefreshToken string
	SessionID    string
	PublicID     string
}

type AuthUser struct {
	ID          string   `json:"id"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions,omitempty"`
}

type AuthResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int64     `json:"expires_in"`
	User        *AuthUser `json:"user"`
}

type MeResponse struct {
	User      AuthUser `json:"user"`
	SessionID string   `json:"session_id"`
}
