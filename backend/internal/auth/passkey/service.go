package passkey

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/storage"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

type Service struct {
	store storage.Provider
	wa    *webauthn.WebAuthn
}

func NewService(store storage.Provider, rpName, rpID, origin string) (*Service, error) {
	wa, err := webauthn.New(&webauthn.Config{
		RPDisplayName: rpName,
		RPID:          rpID,
		RPOrigins:     []string{origin},
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			ResidentKey: protocol.ResidentKeyRequirementRequired,
		},
		Timeouts: webauthn.TimeoutsConfig{
			Login:        webauthn.TimeoutConfig{Timeout: 120 * time.Second},
			Registration: webauthn.TimeoutConfig{Timeout: 120 * time.Second},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("init webauthn: %w", err)
	}
	return &Service{store: store, wa: wa}, nil
}

func (s *Service) BeginRegistration(ctx context.Context, userID string) (*protocol.CredentialCreation, string, error) {
	user, err := s.store.GetUser(ctx, userID)
	if err != nil {
		return nil, "", err
	}

	existingCreds, _ := s.store.ListPasskeysByUser(ctx, userID)
	waUser := NewWebAuthnUser(user, existingCreds)

	var opts []webauthn.RegistrationOption
	if len(existingCreds) > 0 {
		excludeCreds := make([]protocol.CredentialDescriptor, 0, len(existingCreds))
		for _, c := range existingCreds {
			excludeCreds = append(excludeCreds, protocol.CredentialDescriptor{
				Type:         protocol.PublicKeyCredentialType,
				CredentialID: decodeStoredBytes(c.CredentialID),
			})
		}
		opts = append(opts, webauthn.WithExclusions(excludeCreds))
	}

	cc, sessionData, err := s.wa.BeginRegistration(waUser, opts...)
	if err != nil {
		return nil, "", err
	}

	sessionJSON, _ := json.Marshal(sessionData)
	challengeID := randomID("chal")
	if err := s.store.CreateChallenge(ctx, storage.ChallengeRecord{
		ID:        challengeID,
		Challenge: encodeStoredBytes(sessionJSON),
		UserID:    userID,
		FlowType:  "passkey_register",
		ExpiresAt: time.Now().UTC().Add(5 * time.Minute),
	}); err != nil {
		return nil, "", err
	}

	return cc, challengeID, nil
}

func (s *Service) FinishRegistration(ctx context.Context, userID, challengeID string, response *protocol.ParsedCredentialCreationData) error {
	chal, err := s.store.GetChallenge(ctx, challengeID)
	if err != nil {
		return fmt.Errorf("challenge not found or expired")
	}
	_ = s.store.DeleteChallenge(ctx, challengeID)

	var sessionData webauthn.SessionData
	sessionJSON := decodeStoredBytes(chal.Challenge)
	if err := json.Unmarshal(sessionJSON, &sessionData); err != nil {
		return fmt.Errorf("invalid challenge data")
	}

	user, err := s.store.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	existingCreds, _ := s.store.ListPasskeysByUser(ctx, userID)
	waUser := NewWebAuthnUser(user, existingCreds)

	credential, err := s.wa.CreateCredential(waUser, sessionData, response)
	if err != nil {
		return fmt.Errorf("verify registration: %w", err)
	}

	transportsJSON, _ := json.Marshal(response.Response.Transports)
	credentialRecord := storage.PasskeyCredential{
		ID:              randomID("pk"),
		UserID:          userID,
		CredentialID:    encodeStoredBytes(credential.ID),
		PublicKey:       encodeStoredBytes(credential.PublicKey),
		AttestationType: credential.AttestationType,
		AAGUID:          encodeStoredBytes(credential.Authenticator.AAGUID),
		SignCount:       credential.Authenticator.SignCount,
		Transports:      string(transportsJSON),
		Name:            fmt.Sprintf("Passkey %s", time.Now().Format("2006-01-02")),
		CreatedAt:       time.Now().UTC(),
	}
	if err := s.store.CreatePasskeyCredential(ctx, credentialRecord); err != nil {
		return err
	}
	return s.store.CreateAuthMethod(ctx, storage.AuthMethodRecord{
		ID:             randomID("auth"),
		UserID:         userID,
		ProviderType:   "passkey",
		ProviderUserID: credentialRecord.CredentialID,
		Metadata:       fmt.Sprintf(`{"passkeyId":%q,"name":%q}`, credentialRecord.ID, credentialRecord.Name),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	})
}

func (s *Service) BeginLogin(ctx context.Context) (*protocol.CredentialAssertion, string, error) {
	// Discoverable login - 不需要预先知道用户
	assertion, sessionData, err := s.wa.BeginDiscoverableLogin()
	if err != nil {
		return nil, "", err
	}

	sessionJSON, _ := json.Marshal(sessionData)
	challengeID := randomID("chal")
	if err := s.store.CreateChallenge(ctx, storage.ChallengeRecord{
		ID:        challengeID,
		Challenge: encodeStoredBytes(sessionJSON),
		FlowType:  "passkey_login",
		ExpiresAt: time.Now().UTC().Add(5 * time.Minute),
	}); err != nil {
		return nil, "", err
	}

	return assertion, challengeID, nil
}

func (s *Service) FinishLogin(ctx context.Context, challengeID string, response *protocol.ParsedCredentialAssertionData) (string, error) {
	chal, err := s.store.GetChallenge(ctx, challengeID)
	if err != nil {
		return "", fmt.Errorf("challenge not found or expired")
	}
	_ = s.store.DeleteChallenge(ctx, challengeID)

	var sessionData webauthn.SessionData
	sessionJSON := decodeStoredBytes(chal.Challenge)
	if err := json.Unmarshal(sessionJSON, &sessionData); err != nil {
		return "", fmt.Errorf("invalid challenge data")
	}

	handler := func(rawID, userHandle []byte) (webauthn.User, error) {
		credID := encodeStoredBytes(rawID)
		cred, err := s.store.GetPasskeyByCredentialID(ctx, credID)
		if err != nil {
			return nil, fmt.Errorf("credential not found")
		}
		user, err := s.store.GetUser(ctx, cred.UserID)
		if err != nil {
			return nil, fmt.Errorf("user not found")
		}
		creds, _ := s.store.ListPasskeysByUser(ctx, user.ID)
		return NewWebAuthnUser(user, creds), nil
	}

	credential, err := s.wa.ValidateDiscoverableLogin(handler, sessionData, response)
	if err != nil {
		return "", fmt.Errorf("verify login: %w", err)
	}

	credID := encodeStoredBytes(credential.ID)
	cred, _ := s.store.GetPasskeyByCredentialID(ctx, credID)
	_ = s.store.UpdatePasskeyLastUsed(ctx, cred.ID, credential.Authenticator.SignCount)

	return cred.UserID, nil
}

func (s *Service) ListPasskeys(ctx context.Context, userID string) ([]storage.PasskeyCredential, error) {
	return s.store.ListPasskeysByUser(ctx, userID)
}

func (s *Service) DeletePasskey(ctx context.Context, userID, passkeyID string) error {
	return s.store.DeletePasskey(ctx, passkeyID, userID)
}

func randomID(prefix string) string {
	raw := make([]byte, 8)
	_, _ = rand.Read(raw)
	return fmt.Sprintf("%s_%x", prefix, raw)
}
