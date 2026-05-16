package api

import (
	"log"
	"net/http"

	"github.com/IsKenKenYa/Commory/backend/internal/auth"
)

type authResponse struct {
	User  auth.UserInfo `json:"user"`
	Token string        `json:"token"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

func writePublicAuthError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch err {
	case auth.ErrInvalidRequest,
		auth.ErrUsernamePasswordMissing,
		auth.ErrRefreshTokenRequired,
		auth.ErrPasskeyInvalidResponse:
		status = http.StatusBadRequest
	case auth.ErrInvalidCredentials,
		auth.ErrTooManyLoginAttempts,
		auth.ErrUnauthorized,
		auth.ErrRefreshTokenExpired,
		auth.ErrRefreshTokenReplayed,
		auth.ErrPasskeyChallengeExpired,
		auth.ErrPasskeyVerification:
		status = http.StatusUnauthorized
	case auth.ErrRefreshTokenRetry:
		status = http.StatusConflict
	case auth.ErrUsernameExists:
		status = http.StatusConflict
	}
	writeError(w, status, auth.PublicErrorCode(err))
}

func logAuthInternalError(scope string, err error) {
	if err == nil {
		return
	}
	log.Printf("[auth] %s: %v", scope, err)
}
