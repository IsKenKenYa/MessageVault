package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
	sqlc "github.com/IsKenKenYa/Commory/backend/internal/storage/sqlc/gen"
)

func (s *sqliteProvider) rowToUser(row *sqlc.User) UserRecord {
	var roles, buttons []string
	json.Unmarshal([]byte(row.Roles), &roles)
	json.Unmarshal([]byte(row.Buttons), &buttons)
	return UserRecord{
		ID:           row.ID,
		UserName:     row.UserName,
		Email:        row.Email.String,
		PasswordHash: row.PasswordHash,
		PasswordSalt: row.PasswordSalt,
		Roles:        roles,
		Buttons:      buttons,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func (s *sqliteProvider) rowToRefreshToken(row *sqlc.RefreshToken) RefreshTokenRecord {
	return RefreshTokenRecord{
		ID:        row.ID,
		UserID:    row.UserID,
		TokenHash: row.TokenHash,
		ParentID:  row.ParentID.String,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
		RevokedAt: row.RevokedAt.Time,
	}
}

func (s *sqliteProvider) rowToSession(row *sqlc.Session) SessionRecord {
	return SessionRecord{
		ID:             row.ID,
		UserID:         row.UserID,
		RefreshTokenID: row.RefreshTokenID.String,
		DeviceName:     row.DeviceName.String,
		DeviceType:     row.DeviceType.String,
		IPAddress:      row.IpAddress.String,
		UserAgent:      row.UserAgent.String,
		CreatedAt:      row.CreatedAt,
		LastSeenAt:     row.LastSeenAt,
	}
}

func (s *sqliteProvider) rowToPasskey(row *sqlc.PasskeyCredential) PasskeyCredential {
	return PasskeyCredential{
		ID:              row.ID,
		UserID:          row.UserID,
		CredentialID:    row.CredentialID,
		PublicKey:       row.PublicKey,
		AttestationType: row.AttestationType,
		AAGUID:          row.Aaguid,
		SignCount:       uint32(row.SignCount),
		Transports:      row.Transports,
		Name:            row.Name,
		LastUsedAt:      row.LastUsedAt.Time,
		CreatedAt:       row.CreatedAt,
	}
}

func (s *sqliteProvider) rowToIdentity(row *sqlc.Identity) msglayer.Identity {
	var phones, emails, labels []string
	var meta map[string]any
	json.Unmarshal([]byte(row.Phones), &phones)
	json.Unmarshal([]byte(row.Emails), &emails)
	json.Unmarshal([]byte(row.Labels), &labels)
	json.Unmarshal([]byte(row.Meta), &meta)
	var avatar *string
	if row.Avatar.Valid {
		avatar = &row.Avatar.String
	}
	return msglayer.Identity{
		ID:          row.ID,
		Type:        row.Type,
		DisplayName: row.DisplayName,
		Phones:      phones,
		Emails:      emails,
		Avatar:      avatar,
		Labels:      labels,
		Meta:        meta,
	}
}

func (s *sqliteProvider) eventToTimelineItem(ctx context.Context, row *sqlc.Event) (msglayer.TimelineItem, error) {
	participants, _ := s.q.GetEventParticipants(ctx, row.ID)
	var meta map[string]any
	json.Unmarshal([]byte(row.Meta), &meta)
	return msglayer.TimelineItem{
		EventID:        row.ID,
		Type:           row.Type,
		Timestamp:      row.Timestamp.UTC().Format(time.RFC3339),
		Direction:      row.Direction,
		ContentSummary: row.ContentSummary.String,
		Participants:   participants,
		Meta:           meta,
	}, nil
}

func (s *sqliteProvider) eventsToTimelineItems(ctx context.Context, rows []*sqlc.Event) ([]msglayer.TimelineItem, error) {
	items := make([]msglayer.TimelineItem, 0, len(rows))
	for _, row := range rows {
		item, err := s.eventToTimelineItem(ctx, row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
