package auth

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

const (
	accessTokenTTL  = 15 * time.Minute
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
	return s.RegisterWithDevice(ctx, userName, email, password, "", "", "")
}

func (s *Service) RegisterWithDevice(ctx context.Context, userName, email, password, deviceName, ipAddress, userAgent string) (UserInfo, TokenPair, error) {
	userName = strings.TrimSpace(userName)
	if userName == "" || strings.TrimSpace(password) == "" {
		return UserInfo{}, TokenPair{}, ErrUsernamePasswordMissing
	}

	if _, err := s.store.FindUserByUserName(ctx, userName); err == nil {
		return UserInfo{}, TokenPair{}, ErrUsernameExists
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

	if err := s.store.CreateAuthMethod(ctx, storage.AuthMethodRecord{
		ID:             randomID("auth"),
		UserID:         record.ID,
		ProviderType:   "password",
		ProviderUserID: record.ID,
		Metadata:       fmt.Sprintf(`{"userName":%q}`, record.UserName),
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		return UserInfo{}, TokenPair{}, err
	}

	pair, _, err := s.issueTokenPairWithSession(ctx, record, "", "", deviceName, ipAddress, userAgent)
	if err != nil {
		return UserInfo{}, TokenPair{}, err
	}

	_ = s.writeAuditLog(ctx, record.ID, "register", ipAddress, userAgent, "")

	return toUserInfo(record), pair, nil
}

func (s *Service) RegisterAdmin(ctx context.Context, userName, email, password string) (UserInfo, TokenPair, error) {
	return s.RegisterAdminWithDevice(ctx, userName, email, password, "", "", "")
}

func (s *Service) RegisterAdminWithDevice(ctx context.Context, userName, email, password, deviceName, ipAddress, userAgent string) (UserInfo, TokenPair, error) {
	userName = strings.TrimSpace(userName)
	if userName == "" || strings.TrimSpace(password) == "" {
		return UserInfo{}, TokenPair{}, ErrUsernamePasswordMissing
	}

	if _, err := s.store.FindUserByUserName(ctx, userName); err == nil {
		return UserInfo{}, TokenPair{}, ErrUsernameExists
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

	if err := s.store.CreateAuthMethod(ctx, storage.AuthMethodRecord{
		ID:             randomID("auth"),
		UserID:         record.ID,
		ProviderType:   "password",
		ProviderUserID: record.ID,
		Metadata:       fmt.Sprintf(`{"userName":%q}`, record.UserName),
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		return UserInfo{}, TokenPair{}, err
	}

	pair, _, err := s.issueTokenPairWithSession(ctx, record, "", "", deviceName, ipAddress, userAgent)
	if err != nil {
		return UserInfo{}, TokenPair{}, err
	}

	_ = s.writeAuditLog(ctx, record.ID, "register", ipAddress, userAgent, "")

	return toUserInfo(record), pair, nil
}

func (s *Service) Login(ctx context.Context, remoteAddr, userName, password string) (UserInfo, TokenPair, error) {
	return s.LoginWithDevice(ctx, remoteAddr, userName, password, "", "", "")
}

func (s *Service) LoginWithDevice(ctx context.Context, remoteAddr, userName, password, deviceName, ipAddress, userAgent string) (UserInfo, TokenPair, error) {
	if err := s.allowLoginAttempt(remoteAddr); err != nil {
		return UserInfo{}, TokenPair{}, err
	}

	record, err := s.store.FindUserByUserName(ctx, strings.TrimSpace(userName))
	if err != nil {
		return UserInfo{}, TokenPair{}, ErrInvalidCredentials
	}
	if !verifyPassword(password, record.PasswordSalt, record.PasswordHash) {
		return UserInfo{}, TokenPair{}, ErrInvalidCredentials
	}

	if NeedsRehash(record.PasswordHash) {
		if newSalt, newHash, rehashErr := hashPassword(password); rehashErr == nil {
			_ = s.store.UpdateUserPasswordHash(ctx, record.ID, newHash, newSalt)
		}
	}

	pair, _, err := s.issueTokenPairWithSession(ctx, record, "", "", deviceName, ipAddress, userAgent)
	if err != nil {
		return UserInfo{}, TokenPair{}, err
	}

	_ = s.writeAuditLog(ctx, record.ID, "login", ipAddress, userAgent, `{"method":"password"}`)

	return toUserInfo(record), pair, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	token := strings.TrimSpace(refreshToken)
	if token == "" {
		return TokenPair{}, ErrRefreshTokenRequired
	}
	tokenHash := hashToken(token)

	record, err := s.store.ConsumeRefreshToken(ctx, tokenHash)
	if err != nil {
		return TokenPair{}, s.classifyRefreshFailure(ctx, tokenHash)
	}

	user, err := s.store.GetUser(ctx, record.UserID)
	if err != nil {
		return TokenPair{}, err
	}

	var (
		sessionID  string
		deviceName string
		ipAddress  string
		userAgent  string
	)
	if session, sessionErr := s.store.GetSessionByRefreshTokenID(ctx, record.ID); sessionErr == nil {
		sessionID = session.ID
		deviceName = session.DeviceName
		ipAddress = session.IPAddress
		userAgent = session.UserAgent
	}

	pair, _, err := s.issueTokenPairWithSession(ctx, user, record.ID, sessionID, deviceName, ipAddress, userAgent)
	if err != nil {
		return TokenPair{}, err
	}

	_ = s.writeAuditLog(ctx, user.ID, "token_refresh", "", "", "")

	return pair, nil
}

func (s *Service) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	record, err := s.store.ConsumeRefreshToken(ctx, hashToken(refreshToken))
	if err != nil {
		return err
	}
	if session, sessionErr := s.store.GetSessionByRefreshTokenID(ctx, record.ID); sessionErr == nil {
		_ = s.store.RevokeSession(ctx, session.ID)
	}
	_ = s.writeAuditLog(ctx, record.UserID, "logout", "", "", "")
	return nil
}

func (s *Service) UserInfo(ctx context.Context, userID string) (UserInfo, error) {
	user, err := s.store.GetUser(ctx, userID)
	if err != nil {
		return UserInfo{}, err
	}
	return toUserInfo(user), nil
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
