package storage

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
)

func EnsureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func (s *fileStore) persist() error {
	data, err := json.MarshalIndent(s.snapshot, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func scopeKey(userID, rawID string) string {
	hash := sha1.Sum([]byte(userID))
	return fmt.Sprintf("%x:%s", hash[:8], rawID)
}

func scopeKeyLegacy(userID, rawID string) string {
	hash := sha1.Sum([]byte(userID))
	return fmt.Sprintf("%x:%s", hash[:4], rawID)
}

func (s *fileStore) lookupScopedKey(m map[string]storedIdentity, userID, rawID string) (storedIdentity, bool) {
	key := scopeKey(userID, rawID)
	if v, ok := m[key]; ok && v.UserID == userID {
		return v, true
	}
	key = scopeKeyLegacy(userID, rawID)
	if v, ok := m[key]; ok && v.UserID == userID {
		return v, true
	}
	return storedIdentity{}, false
}

func (s *fileStore) lookupScopedEventKey(m map[string]storedEvent, userID, rawID string) (storedEvent, bool) {
	key := scopeKey(userID, rawID)
	if v, ok := m[key]; ok && v.UserID == userID {
		return v, true
	}
	key = scopeKeyLegacy(userID, rawID)
	if v, ok := m[key]; ok && v.UserID == userID {
		return v, true
	}
	return storedEvent{}, false
}

func summarizeEvent(event msglayer.Event) string {
	switch event.Type {
	case "sms":
		return asString(event.Content["text"])
	case "call":
		return fmt.Sprintf("%s %vsec", asString(event.Content["call_type"]), event.Content["duration_sec"])
	case "voice":
		if summary := asString(event.Content["summary"]); summary != "" {
			return summary
		}
		return asString(event.Content["transcript"])
	case "contact_snapshot":
		return asString(event.Content["identity_id"])
	default:
		return event.Type
	}
}

func matchesKeyword(event storedEvent, keyword string) bool {
	needle := strings.ToLower(keyword)
	if strings.Contains(strings.ToLower(event.Item.ContentSummary), needle) {
		return true
	}
	if transcript := asString(event.Raw.Content["transcript"]); strings.Contains(strings.ToLower(transcript), needle) {
		return true
	}
	if summary := asString(event.Raw.Content["summary"]); strings.Contains(strings.ToLower(summary), needle) {
		return true
	}
	return false
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	if value, ok := v.(string); ok {
		return value
	}
	return fmt.Sprintf("%v", v)
}
