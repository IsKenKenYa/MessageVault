package auth

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

func detectDeviceType(userAgent string) string {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "commory"):
		return "android"
	case strings.Contains(ua, "mozilla") || strings.Contains(ua, "chrome") || strings.Contains(ua, "safari"):
		return "web"
	default:
		return "unknown"
	}
}

func (s *Service) writeAuditLog(ctx context.Context, userID, action, ipAddress, userAgent, detail string) error {
	return s.store.CreateAuditLog(ctx, storage.AuditRecord{
		ID:        randomID("audit"),
		UserID:    userID,
		Action:    action,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Detail:    detail,
		CreatedAt: time.Now().UTC(),
	})
}

func (s *Service) WriteAuditLog(ctx context.Context, userID, action, ipAddress, userAgent, detail string) error {
	return s.writeAuditLog(ctx, userID, action, ipAddress, userAgent, detail)
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
		return ErrTooManyLoginAttempts
	}
	s.loginWindow[host] = append(window, now)
	return nil
}
