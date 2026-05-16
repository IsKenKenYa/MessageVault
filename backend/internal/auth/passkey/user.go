package passkey

import (
	"github.com/IsKenKenYa/Commory/backend/internal/storage"
	"github.com/go-webauthn/webauthn/webauthn"
)

type WebAuthnUser struct {
	user        storage.UserRecord
	credentials []webauthn.Credential
}

func NewWebAuthnUser(user storage.UserRecord, creds []storage.PasskeyCredential) *WebAuthnUser {
	waCreds := make([]webauthn.Credential, 0, len(creds))
	for _, c := range creds {
		waCred := webauthn.Credential{
			ID:              []byte(c.CredentialID),
			PublicKey:        []byte(c.PublicKey),
			AttestationType: c.AttestationType,
			Authenticator: webauthn.Authenticator{
				AAGUID:       []byte(c.AAGUID),
				SignCount:    c.SignCount,
				CloneWarning: false,
			},
		}
		waCreds = append(waCreds, waCred)
	}
	return &WebAuthnUser{user: user, credentials: waCreds}
}

func (u *WebAuthnUser) WebAuthnID() []byte {
	return []byte(u.user.ID)
}

func (u *WebAuthnUser) WebAuthnName() string {
	return u.user.UserName
}

func (u *WebAuthnUser) WebAuthnDisplayName() string {
	if u.user.Email != "" {
		return u.user.Email
	}
	return u.user.UserName
}

func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

func (u *WebAuthnUser) WebAuthnIcon() string {
	return ""
}
