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

func (s Service) Event(ctx context.Context, id string) (msglayer.TimelineItem, error) {
	return s.store.GetEvent(ctx, id)
}

func (s Service) Thread(ctx context.Context, id string) ([]msglayer.TimelineItem, error) {
	return s.store.GetThread(ctx, id)
}

func (s Service) Identities(ctx context.Context) ([]msglayer.Identity, error) {
	return s.store.ListIdentities(ctx)
}

func (s Service) Identity(ctx context.Context, id string) (msglayer.Identity, error) {
	return s.store.GetIdentity(ctx, id)
}
