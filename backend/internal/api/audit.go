package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/IsKenKenYa/Commory/backend/internal/auth"
)

func (s *Server) handleAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := s.requireAdmin(r.Context()); err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	q := r.URL.Query()
	userID := q.Get("user_id")
	action := q.Get("action")
	limit := 50
	if v, err := strconv.Atoi(q.Get("limit")); err == nil && v > 0 && v <= 200 {
		limit = v
	}
	offset := 0
	if v, err := strconv.Atoi(q.Get("offset")); err == nil && v >= 0 {
		offset = v
	}

	logs, err := s.store.ListAuditLogs(r.Context(), userID, action, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	total, _ := s.store.CountAuditLogs(r.Context(), userID, action)
	writeJSON(w, http.StatusOK, "ok", map[string]any{
		"items":  logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (s *Server) requireAdmin(ctx context.Context) error {
	user, err := s.auth.UserInfo(ctx, auth.UserIDFromContext(ctx))
	if err != nil {
		return err
	}
	for _, role := range user.Roles {
		if role == "R_SUPER" || role == "R_ADMIN" {
			return nil
		}
	}
	return fmt.Errorf("admin privileges required")
}
