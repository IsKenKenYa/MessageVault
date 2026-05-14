package msglayer

const Version = "msglayer/v0.1"

type RootExport struct {
	Version    string         `json:"version"`
	ExportedAt string         `json:"exported_at"`
	Source     Source         `json:"source"`
	Identities []Identity     `json:"identities"`
	Events     []Event        `json:"events"`
	Indexes    map[string]any `json:"indexes,omitempty"`
}

type Source struct {
	Platform   string `json:"platform"`
	DeviceID   string `json:"device_id"`
	AppVersion string `json:"app_version"`
}

type Identity struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	DisplayName string         `json:"display_name"`
	Phones      []string       `json:"phones"`
	Emails      []string       `json:"emails"`
	Avatar      *string        `json:"avatar"`
	Labels      []string       `json:"labels"`
	Meta        map[string]any `json:"meta"`
}

type Event struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Timestamp    string         `json:"timestamp"`
	Direction    string         `json:"direction"`
	Participants []string       `json:"participants"`
	Content      map[string]any `json:"content"`
	Meta         map[string]any `json:"meta"`
	Relations    []Relation     `json:"relations"`
}

type Relation struct {
	Type   string `json:"type"`
	Target string `json:"target"`
}

type SearchParams struct {
	UserID      string
	Keyword     string
	ContactID   string
	Type        string
	Participant string
	From        string
	To          string
	Limit       int
}

type TimelineItem struct {
	EventID        string         `json:"event_id"`
	Type           string         `json:"type"`
	Timestamp      string         `json:"timestamp"`
	Direction      string         `json:"direction"`
	ContentSummary string         `json:"content_summary"`
	Participants   []string       `json:"participants"`
	Meta           map[string]any `json:"meta,omitempty"`
	SchemaVersion  string         `json:"schema_version"`
}
