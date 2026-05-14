package query

import (
	"context"

	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

type Service struct {
	store storage.Provider
}

func New(store storage.Provider) Service {
	return Service{store: store}
}

func (s Service) Events(ctx context.Context, params msglayer.SearchParams) ([]msglayer.TimelineItem, error) {
	return s.store.ListEvents(ctx, params)
}

func (s Service) Search(ctx context.Context, params msglayer.SearchParams) ([]msglayer.TimelineItem, error) {
	return s.store.Search(ctx, params)
}

func (s Service) Timeline(ctx context.Context, params msglayer.SearchParams) ([]msglayer.TimelineItem, error) {
	return s.store.Timeline(ctx, params)
}

func (s Service) Event(ctx context.Context, userID, id string) (msglayer.TimelineItem, error) {
	return s.store.GetEvent(ctx, userID, id)
}

func (s Service) Thread(ctx context.Context, userID, id string) ([]msglayer.TimelineItem, error) {
	return s.store.GetThread(ctx, userID, id)
}

func (s Service) Identities(ctx context.Context, userID string) ([]msglayer.Identity, error) {
	return s.store.ListIdentities(ctx, userID)
}

func (s Service) Identity(ctx context.Context, userID, id string) (msglayer.Identity, error) {
	return s.store.GetIdentity(ctx, userID, id)
}
