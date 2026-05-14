package storage

import (
	"context"

	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
)

type Provider interface {
	Name() string
	Init(context.Context) error
	Import(context.Context, string, msglayer.RootExport, []byte) (string, error)
	ExportImport(context.Context, string) ([]byte, error)
	LatestImportID(context.Context) (string, error)
	GetEvent(context.Context, string) (msglayer.TimelineItem, error)
	ListEvents(context.Context, msglayer.SearchParams) ([]msglayer.TimelineItem, error)
	Search(context.Context, msglayer.SearchParams) ([]msglayer.TimelineItem, error)
	Timeline(context.Context, msglayer.SearchParams) ([]msglayer.TimelineItem, error)
	ListIdentities(context.Context) ([]msglayer.Identity, error)
	GetIdentity(context.Context, string) (msglayer.Identity, error)
	GetThread(context.Context, string) ([]msglayer.TimelineItem, error)
	Close() error
}
