package importers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
)

type Importer interface {
	Import(path string) (msglayer.RootExport, []byte, error)
	ImportBytes([]byte) (msglayer.RootExport, []byte, error)
}

type JSONImporter struct{}

func (JSONImporter) Import(path string) (msglayer.RootExport, []byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return msglayer.RootExport{}, nil, err
	}
	return JSONImporter{}.ImportBytes(data)
}

func (JSONImporter) ImportBytes(data []byte) (msglayer.RootExport, []byte, error) {
	var export msglayer.RootExport
	if err := json.Unmarshal(data, &export); err != nil {
		return msglayer.RootExport{}, nil, fmt.Errorf("decode msglayer json: %w", err)
	}
	return export, data, nil
}
