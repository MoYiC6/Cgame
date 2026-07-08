package auth

type LoginRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Platform   string `json:"platform"`
	LoginType  string `json:"loginType"`
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
	AccessToken    string    `json:"access_token"`
	RefreshToken   string    `json:"refresh_token"`
	TokenType      string    `json:"token_type"`
	ExpiresIn      int64     `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	UserID         int64     `json:"user_id"`
	Username       string    `json:"username"`
	Nickname       string    `json:"nickname"`
	Avatar         string    `json:"avatar"`
	Roles          string    `json:"roles"`
	Permissions    []string  `json:"permissions"`
	ProfileIncomplete bool  `json:"profile_incomplete"`
	IncompleteFields []string `json:"incomplete_fields"`
}

type MeResponse struct {
	UserID      int64     `json:"userId"`
	Username    string    `json:"username"`
	Nickname    string    `json:"nickname"`
	Email       string    `json:"email"`
	Avatar      string    `json:"avatar"`
	Gender      int16     `json:"gender"`
	Mobile      string    `json:"mobile"`
	IsTeacher   int16     `json:"isTeacher"`
	TeacherID   *int64    `json:"teacherId"`
	Status      string    `json:"status"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	Menus       []any     `json:"menus"`
	Buttons     []string  `json:"buttons"`
	SessionID   string    `json:"session_id"`
}
