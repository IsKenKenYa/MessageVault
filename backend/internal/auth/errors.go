package auth

type codeError string

func (e codeError) Error() string {
	return string(e)
}

const (
	ErrInvalidRequest          = codeError("ERR_INVALID_REQUEST")
	ErrUnauthorized            = codeError("ERR_UNAUTHORIZED")
	ErrInvalidCredentials      = codeError("ERR_INVALID_CREDENTIALS")
	ErrUsernameExists          = codeError("ERR_USERNAME_EXISTS")
	ErrUsernamePasswordMissing = codeError("ERR_USERNAME_PASSWORD_REQUIRED")
	ErrTooManyLoginAttempts    = codeError("ERR_TOO_MANY_LOGIN_ATTEMPTS")
	ErrRefreshTokenRequired    = codeError("ERR_REFRESH_TOKEN_REQUIRED")
	ErrRefreshTokenExpired     = codeError("ERR_REFRESH_TOKEN_EXPIRED")
	ErrRefreshTokenReplayed    = codeError("ERR_REFRESH_TOKEN_REPLAYED")
	ErrRefreshTokenRetry       = codeError("ERR_REFRESH_TOKEN_RETRY")
	ErrPasskeyChallengeExpired = codeError("ERR_PASSKEY_CHALLENGE_EXPIRED")
	ErrPasskeyInvalidResponse  = codeError("ERR_PASSKEY_INVALID_RESPONSE")
	ErrPasskeyVerification     = codeError("ERR_PASSKEY_VERIFICATION_FAILED")
	ErrOperationFailed         = codeError("ERR_OPERATION_FAILED")
)

func PublicErrorCode(err error) string {
	switch err {
	case nil:
		return ""
	case ErrInvalidRequest,
		ErrUnauthorized,
		ErrInvalidCredentials,
		ErrUsernameExists,
		ErrUsernamePasswordMissing,
		ErrTooManyLoginAttempts,
		ErrRefreshTokenRequired,
		ErrRefreshTokenExpired,
		ErrRefreshTokenReplayed,
		ErrRefreshTokenRetry,
		ErrPasskeyChallengeExpired,
		ErrPasskeyInvalidResponse,
		ErrPasskeyVerification:
		return err.Error()
	default:
		return ErrOperationFailed.Error()
	}
}
