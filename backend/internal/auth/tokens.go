package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

type AccessTokenClaims struct {
	UserID    string
	SessionID string
	ExpiresAt int64
}

// IssueTokenPairForUser 为已验证的用户签发 token pair（Passkey 登录用）
func (s *Service) IssueTokenPairForUser(ctx context.Context, user storage.UserRecord, deviceName, ipAddress, userAgent string) (TokenPair, string, error) {
	return s.issueTokenPairWithSession(ctx, user, "", "", deviceName, ipAddress, userAgent)
}

func (s *Service) ParseAccessToken(token string) (string, error) {
	claims, err := s.ParseAccessTokenClaims(token)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

func (s *Service) ParseAccessTokenClaims(token string) (AccessTokenClaims, error) {
	token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return AccessTokenClaims{}, fmt.Errorf("invalid token")
	}

	signingInput := strings.Join(parts[:2], ".")
	expected := signHS256(signingInput, s.secret)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return AccessTokenClaims{}, fmt.Errorf("invalid token signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return AccessTokenClaims{}, fmt.Errorf("decode token payload: %w", err)
	}
	var payload struct {
		Sub string `json:"sub"`
		Sid string `json:"sid"`
		Exp int64  `json:"exp"`
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return AccessTokenClaims{}, fmt.Errorf("decode token json: %w", err)
	}
	if payload.Sub == "" || time.Now().Unix() > payload.Exp {
		return AccessTokenClaims{}, fmt.Errorf("token expired")
	}
	return AccessTokenClaims{
		UserID:    payload.Sub,
		SessionID: payload.Sid,
		ExpiresAt: payload.Exp,
	}, nil
}

func (s *Service) issueTokenPairWithSession(ctx context.Context, user storage.UserRecord, parentID, sessionID, deviceName, ipAddress, userAgent string) (TokenPair, string, error) {
	refreshToken, refreshHash, err := buildRefreshToken()
	if err != nil {
		return TokenPair{}, "", err
	}
	refreshID := randomID("refresh")
	if sessionID == "" {
		sessionID = randomID("session")
	}
	accessToken, err := s.buildAccessToken(user, sessionID)
	if err != nil {
		return TokenPair{}, "", err
	}
	if err := s.store.SaveRefreshToken(ctx, storage.RefreshTokenRecord{
		ID:        refreshID,
		UserID:    user.ID,
		TokenHash: refreshHash,
		ParentID:  parentID,
		ExpiresAt: time.Now().UTC().Add(RefreshTokenTTL),
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		return TokenPair{}, "", err
	}

	now := time.Now().UTC()
	if parentID == "" {
		if err := s.store.CreateSession(ctx, storage.SessionRecord{
			ID:             sessionID,
			UserID:         user.ID,
			RefreshTokenID: refreshID,
			DeviceName:     deviceName,
			DeviceType:     detectDeviceType(userAgent),
			IPAddress:      ipAddress,
			UserAgent:      userAgent,
			CreatedAt:      now,
			LastSeenAt:     now,
		}); err != nil {
			return TokenPair{}, "", err
		}
	} else {
		if err := s.store.UpdateSessionRefreshToken(ctx, sessionID, refreshID); err != nil {
			return TokenPair{}, "", err
		}
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, sessionID, nil
}

func (s *Service) buildAccessToken(user storage.UserRecord, sessionID string) (string, error) {
	headerBytes, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payloadBytes, err := json.Marshal(map[string]any{
		"sub":      user.ID,
		"sid":      sessionID,
		"userName": user.UserName,
		"roles":    user.Roles,
		"exp":      time.Now().Add(accessTokenTTL).Unix(),
	})
	if err != nil {
		return "", err
	}
	header := base64.RawURLEncoding.EncodeToString(headerBytes)
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := header + "." + payload
	return signingInput + "." + signHS256(signingInput, s.secret), nil
}

func buildRefreshToken() (token string, tokenHash string, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	return token, hashToken(token), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func signHS256(input string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *Service) classifyRefreshFailure(ctx context.Context, tokenHash string) error {
	record, err := s.store.FindAnyRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return ErrUnauthorized
	}
	if !record.RevokedAt.IsZero() {
		hasActiveChild, childErr := s.store.HasActiveRefreshTokenChild(ctx, record.ID)
		if childErr != nil {
			return ErrOperationFailed
		}
		if hasActiveChild {
			return ErrRefreshTokenRetry
		}
		_ = s.store.RevokeRefreshTokenFamily(ctx, record.ID)
		return ErrRefreshTokenReplayed
	}
	if time.Now().UTC().After(record.ExpiresAt) {
		return ErrRefreshTokenExpired
	}
	return ErrUnauthorized
}

func randomID(prefix string) string {
	raw := make([]byte, 8)
	_, _ = rand.Read(raw)
	return fmt.Sprintf("%s_%x", prefix, raw)
}
