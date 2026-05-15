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
	"net"
	"strings"
	"sync"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

const (
	accessTokenTTL  = 30 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

type Service struct {
	store       storage.Provider
	secret      []byte
	loginMu     sync.Mutex
	loginWindow map[string][]time.Time
}

func NewService(store storage.Provider, secret string) *Service {
	return &Service{
		store:       store,
		secret:      []byte(secret),
		loginWindow: map[string][]time.Time{},
	}
}

func (s *Service) Register(ctx context.Context, userName, email, password string) (UserInfo, TokenPair, error) {
	userName = strings.TrimSpace(userName)
	if userName == "" || strings.TrimSpace(password) == "" {
		return UserInfo{}, TokenPair{}, fmt.Errorf("username and password are required")
	}

	if _, err := s.store.FindUserByUserName(ctx, userName); err == nil {
		return UserInfo{}, TokenPair{}, fmt.Errorf("username already exists")
	}

	salt, hashed, err := hashPassword(password)
	if err != nil {
		return UserInfo{}, TokenPair{}, err
	}

	roles := []string{"R_USER"}
	buttons := []string{"view", "import"}

	now := time.Now().UTC()
	record, err := s.store.CreateUser(ctx, storage.UserRecord{
		ID:           randomID("user"),
		UserName:     userName,
		Email:        strings.TrimSpace(email),
		PasswordHash: hashed,
		PasswordSalt: salt,
		Roles:        roles,
		Buttons:      buttons,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return UserInfo{}, TokenPair{}, err
	}

	pair, err := s.issueTokenPair(ctx, record)
	if err != nil {
		return UserInfo{}, TokenPair{}, err
	}
	return toUserInfo(record), pair, nil
}

func (s *Service) RegisterAdmin(ctx context.Context, userName, email, password string) (UserInfo, TokenPair, error) {
	userName = strings.TrimSpace(userName)
	if userName == "" || strings.TrimSpace(password) == "" {
		return UserInfo{}, TokenPair{}, fmt.Errorf("username and password are required")
	}

	if _, err := s.store.FindUserByUserName(ctx, userName); err == nil {
		return UserInfo{}, TokenPair{}, fmt.Errorf("username already exists")
	}

	salt, hashed, err := hashPassword(password)
	if err != nil {
		return UserInfo{}, TokenPair{}, err
	}

	roles := []string{"R_ADMIN", "R_USER"}
	buttons := []string{"view", "import", "admin"}

	now := time.Now().UTC()
	record, err := s.store.CreateUser(ctx, storage.UserRecord{
		ID:           randomID("user"),
		UserName:     userName,
		Email:        strings.TrimSpace(email),
		PasswordHash: hashed,
		PasswordSalt: salt,
		Roles:        roles,
		Buttons:      buttons,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return UserInfo{}, TokenPair{}, err
	}

	pair, err := s.issueTokenPair(ctx, record)
	if err != nil {
		return UserInfo{}, TokenPair{}, err
	}
	return toUserInfo(record), pair, nil
}

func (s *Service) Login(ctx context.Context, remoteAddr, userName, password string) (UserInfo, TokenPair, error) {
	if err := s.allowLoginAttempt(remoteAddr); err != nil {
		return UserInfo{}, TokenPair{}, err
	}

	record, err := s.store.FindUserByUserName(ctx, strings.TrimSpace(userName))
	if err != nil {
		return UserInfo{}, TokenPair{}, fmt.Errorf("invalid username or password")
	}
	if !verifyPassword(password, record.PasswordSalt, record.PasswordHash) {
		return UserInfo{}, TokenPair{}, fmt.Errorf("invalid username or password")
	}

	if NeedsRehash(record.PasswordHash) {
		if _, newHash, rehashErr := hashPassword(password); rehashErr == nil {
			_ = s.store.UpdateUserPasswordHash(ctx, record.ID, newHash)
		}
	}

	pair, err := s.issueTokenPair(ctx, record)
	if err != nil {
		return UserInfo{}, TokenPair{}, err
	}
	return toUserInfo(record), pair, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	token := strings.TrimSpace(refreshToken)
	if token == "" {
		return TokenPair{}, fmt.Errorf("refresh token is required")
	}
	record, err := s.store.ConsumeRefreshToken(ctx, hashToken(token))
	if err != nil {
		return TokenPair{}, err
	}
	user, err := s.store.GetUser(ctx, record.UserID)
	if err != nil {
		return TokenPair{}, err
	}
	return s.issueTokenPair(ctx, user)
}

func (s *Service) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	_, err := s.store.ConsumeRefreshToken(ctx, hashToken(refreshToken))
	return err
}

func (s *Service) UserInfo(ctx context.Context, userID string) (UserInfo, error) {
	user, err := s.store.GetUser(ctx, userID)
	if err != nil {
		return UserInfo{}, err
	}
	return toUserInfo(user), nil
}

func (s *Service) ParseAccessToken(token string) (string, error) {
	token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token")
	}

	signingInput := strings.Join(parts[:2], ".")
	expected := signHS256(signingInput, s.secret)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return "", fmt.Errorf("invalid token signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("decode token payload: %w", err)
	}
	var payload struct {
		Sub string `json:"sub"`
		Exp int64  `json:"exp"`
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return "", fmt.Errorf("decode token json: %w", err)
	}
	if payload.Sub == "" || time.Now().Unix() > payload.Exp {
		return "", fmt.Errorf("token expired")
	}
	return payload.Sub, nil
}

func (s *Service) issueTokenPair(ctx context.Context, user storage.UserRecord) (TokenPair, error) {
	accessToken, err := s.buildAccessToken(user)
	if err != nil {
		return TokenPair{}, err
	}
	refreshToken, refreshHash, err := buildRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}
	if err := s.store.SaveRefreshToken(ctx, storage.RefreshTokenRecord{
		ID:        randomID("refresh"),
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: time.Now().UTC().Add(refreshTokenTTL),
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) buildAccessToken(user storage.UserRecord) (string, error) {
	headerBytes, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payloadBytes, err := json.Marshal(map[string]any{
		"sub":      user.ID,
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

func (s *Service) allowLoginAttempt(remoteAddr string) error {
	s.loginMu.Lock()
	defer s.loginMu.Unlock()

	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	now := time.Now()
	cutoff := now.Add(-time.Minute)
	window := s.loginWindow[host][:0]
	for _, ts := range s.loginWindow[host] {
		if ts.After(cutoff) {
			window = append(window, ts)
		}
	}
	if len(window) >= 5 {
		s.loginWindow[host] = window
		return fmt.Errorf("too many login attempts, please retry in a minute")
	}
	s.loginWindow[host] = append(window, now)
	return nil
}

func toUserInfo(user storage.UserRecord) UserInfo {
	return UserInfo{
		Buttons:  append([]string(nil), user.Buttons...),
		Roles:    append([]string(nil), user.Roles...),
		UserID:   user.ID,
		UserName: user.UserName,
		Email:    user.Email,
	}
}

func hashPassword(password string) (salt string, hash string, err error) {
	rawSalt := make([]byte, 16)
	if _, err := rand.Read(rawSalt); err != nil {
		return "", "", err
	}
	salt = hex.EncodeToString(rawSalt)
	sum := derivePasswordPBKDF2(password, salt, 120000)
	return salt, "pbkdf2$" + hex.EncodeToString(sum[:]), nil
}

func verifyPassword(password, salt, expected string) bool {
	var sum [32]byte
	if strings.HasPrefix(expected, "pbkdf2$") {
		sum = derivePasswordPBKDF2(password, salt, 120000)
		return hmac.Equal([]byte("pbkdf2$"+hex.EncodeToString(sum[:])), []byte(expected))
	}
	stripped := strings.TrimPrefix(expected, "sha256$")
	sum = derivePasswordLegacy(password, salt)
	return hmac.Equal([]byte(hex.EncodeToString(sum[:])), []byte(stripped))
}

func NeedsRehash(hash string) bool {
	return !strings.HasPrefix(hash, "pbkdf2$")
}

func derivePasswordPBKDF2(password, salt string, iterations int) [32]byte {
	prf := hmac.New(sha256.New, []byte(password))
	var result [32]byte
	prf.Write([]byte(salt))
	prf.Write([]byte{0, 0, 0, 1})
	u := prf.Sum(nil)
	copy(result[:], u)
	for i := 1; i < iterations; i++ {
		prf.Reset()
		prf.Write(u)
		u = prf.Sum(nil)
		for j := 0; j < 32; j++ {
			result[j] ^= u[j]
		}
	}
	return result
}

func derivePasswordLegacy(password, salt string) [32]byte {
	input := []byte(password + ":" + salt)
	sum := sha256.Sum256(input)
	for i := 0; i < 120000; i++ {
		sum = sha256.Sum256(append(sum[:], input...))
	}
	return sum
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

func randomID(prefix string) string {
	raw := make([]byte, 8)
	_, _ = rand.Read(raw)
	return fmt.Sprintf("%s_%x", prefix, raw)
}
