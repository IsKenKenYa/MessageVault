package passkey

import "encoding/base64"

func encodeStoredBytes(raw []byte) string {
	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeStoredBytes(value string) []byte {
	if value == "" {
		return nil
	}
	if decoded, err := base64.RawURLEncoding.DecodeString(value); err == nil {
		return decoded
	}
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
		return decoded
	}
	return []byte(value)
}
