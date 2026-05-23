package security

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenInvalid = errors.New("token invalid")
	ErrTokenExpired = errors.New("token expired")
)

const minHS256SecretBytes = 32

type AccessToken struct {
	Token     string
	TokenType string
	ExpiresIn int64
	ExpiresAt time.Time
}

type TokenClaims struct {
	TokenID     string
	Subject     string
	SessionID   string
	Issuer      string
	Audience    string
	IssuedAt    time.Time
	NotBefore   time.Time
	ExpiresAt   time.Time
	Roles       []string
	Permissions []string
	Status      string
	PublicID    string
}

type TokenManager interface {
	IssueAccessToken(ctx context.Context, p *Principal) (*AccessToken, error)
	VerifyAccessToken(ctx context.Context, raw string) (*Principal, *TokenClaims, error)
}

type HMACTokenConfig struct {
	Issuer         string
	Audience       string
	KeyID          string
	Secret         []byte
	AccessTokenTTL time.Duration
	ClockSkew      time.Duration
}

type customClaims struct {
	SessionID   string   `json:"sid"`
	PublicID    string   `json:"pid,omitempty"`
	Status      string   `json:"status,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

type HMACTokenManager struct {
	config HMACTokenConfig
	random RandomTokenGenerator
}

func NewHMACTokenManager(cfg HMACTokenConfig) *HMACTokenManager {
	if len(cfg.Secret) < minHS256SecretBytes {
		panic("jwt hmac secret must be at least 32 bytes")
	}
	if strings.TrimSpace(cfg.KeyID) == "" {
		panic("jwt key id is required")
	}
	return &HMACTokenManager{config: cfg, random: CryptoRandomTokenGenerator{}}
}

func (m *HMACTokenManager) IssueAccessToken(ctx context.Context, p *Principal) (*AccessToken, error) {
	now := time.Now().UTC()
	jwtID, err := m.random.GenerateURLSafe(16)
	if err != nil {
		return nil, err
	}
	subject := strings.TrimSpace(p.UserID)
	if subject == "" {
		subject = strings.TrimSpace(p.PublicID)
	}
	claims := customClaims{
		SessionID:   p.SessionID,
		PublicID:    strings.TrimSpace(p.PublicID),
		Status:      strings.TrimSpace(p.Status),
		Roles:       NormalizeStrings(p.Roles),
		Permissions: NormalizeStrings(p.Permissions),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			Subject:   subject,
			Audience:  jwt.ClaimStrings{m.config.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.AccessTokenTTL)),
			ID:        jwtID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = m.config.KeyID
	raw, err := token.SignedString(m.config.Secret)
	if err != nil {
		return nil, err
	}
	return &AccessToken{Token: raw, TokenType: "Bearer", ExpiresIn: int64(m.config.AccessTokenTTL.Seconds()), ExpiresAt: now.Add(m.config.AccessTokenTTL)}, nil
}

func (m *HMACTokenManager) VerifyAccessToken(ctx context.Context, raw string) (*Principal, *TokenClaims, error) {
	claims := &customClaims{}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(m.config.Issuer),
		jwt.WithAudience(m.config.Audience),
		jwt.WithLeeway(m.config.ClockSkew),
		jwt.WithIssuedAt(),
		jwt.WithExpirationRequired(),
		jwt.WithNotBeforeRequired(),
	)
	token, err := parser.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrTokenInvalid
		}
		if strings.TrimSpace(fmt.Sprint(token.Header["kid"])) != m.config.KeyID {
			return nil, ErrTokenInvalid
		}
		return m.config.Secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, nil, ErrTokenExpired
		}
		return nil, nil, ErrTokenInvalid
	}
	if !token.Valid {
		return nil, nil, ErrTokenInvalid
	}
	if claims.IssuedAt == nil || claims.NotBefore == nil || claims.ExpiresAt == nil {
		return nil, nil, ErrTokenInvalid
	}
	publicID := strings.TrimSpace(claims.PublicID)
	userID := strings.TrimSpace(claims.Subject)
	if publicID == "" {
		publicID = userID
		userID = ""
	}
	principal := &Principal{
		UserID:      userID,
		PublicID:    publicID,
		SessionID:   claims.SessionID,
		Roles:       NormalizeStrings(claims.Roles),
		Permissions: NormalizeStrings(claims.Permissions),
		Status:      strings.TrimSpace(claims.Status),
	}
	resultClaims := &TokenClaims{
		TokenID:     claims.ID,
		Subject:     claims.Subject,
		SessionID:   claims.SessionID,
		Issuer:      claims.Issuer,
		Audience:    m.config.Audience,
		IssuedAt:    claims.IssuedAt.Time,
		NotBefore:   claims.NotBefore.Time,
		ExpiresAt:   claims.ExpiresAt.Time,
		Roles:       principal.Roles,
		Permissions: principal.Permissions,
		Status:      principal.Status,
		PublicID:    principal.PublicID,
	}
	return principal, resultClaims, nil
}
