package auth

import "time"

type AuthSession struct {
	ID            string
	UserID        int64
	Status        string
	UserAgentHash string
	IPHash        string
	CreatedAt     time.Time
	LastSeenAt    *time.Time
	RevokedAt     *time.Time
	ExpiresAt     time.Time
}

type RefreshToken struct {
	ID                int64
	UserID            int64
	SessionID         string
	TokenHash         string
	FamilyID          string
	ReplacedByTokenID *int64
	RevokedAt         *time.Time
	UsedAt            *time.Time
	ExpiresAt         time.Time
	CreatedAt         time.Time
}

type LoginAttempt struct {
	IdentifierHash string
	Success        bool
	Reason         string
	IPHash         string
	UserAgentHash  string
	RequestID      string
	TraceID        string
	CreatedAt      time.Time
}

type AuditLog struct {
	EventType     string
	Result        string
	UserPublicID  string
	SessionID     string
	RequestID     string
	TraceID       string
	IPHash        string
	UserAgentHash string
	MetadataJSON  map[string]any
	OccurredAt    time.Time
}

type RefreshCookie struct {
	Value     string
	ExpiresAt time.Time
	Clear     bool
}

type ServiceConfig struct {
	RefreshTokenTTL   time.Duration
	RefreshCookieName string
}
