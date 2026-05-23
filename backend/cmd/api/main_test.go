package main

import (
	"strings"
	"testing"
	"time"

	"backend/internal/platform/config"
	"backend/internal/platform/security"
)

func TestNewTokenManagerRejectsUnsupportedAlgorithm(t *testing.T) {
	cfg := &config.Config{}
	cfg.Auth.Issuer = "backend"
	cfg.Auth.Audience = "admin-api"
	cfg.Auth.AccessTokenTTL = 15 * time.Minute
	cfg.Auth.JWT.Algorithm = "EdDSA"
	cfg.Auth.JWT.KeyID = "eddsa-key-1"

	_, err := newTokenManager(cfg)
	if err == nil {
		t.Fatal("expected unsupported algorithm error")
	}
	if !strings.Contains(err.Error(), "unsupported jwt algorithm") {
		t.Fatalf("expected unsupported algorithm error, got %v", err)
	}
}

func TestNewTokenManagerBuildsHMACManagerForHS256(t *testing.T) {
	t.Setenv("JWT_HMAC_SECRET", "01234567890123456789012345678901")
	cfg := &config.Config{}
	cfg.Auth.Issuer = "backend"
	cfg.Auth.Audience = "admin-api"
	cfg.Auth.AccessTokenTTL = 15 * time.Minute
	cfg.Auth.JWT.Algorithm = "HS256"
	cfg.Auth.JWT.KeyID = "test-key"

	manager, err := newTokenManager(cfg)
	if err != nil {
		t.Fatalf("newTokenManager() error = %v", err)
	}
	if _, ok := manager.(*security.HMACTokenManager); !ok {
		t.Fatalf("expected HMACTokenManager, got %T", manager)
	}
}
