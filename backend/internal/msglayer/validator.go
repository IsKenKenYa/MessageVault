package msglayer

import (
	"encoding/json"
	"fmt"
	"os"
)

type Validator struct {
	rootSchemaPath string
}

func NewValidator(rootSchemaPath string) (*Validator, error) {
	return &Validator{rootSchemaPath: rootSchemaPath}, nil
}

func (v *Validator) ValidateBytes(data []byte) error {
	var payload RootExport
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	if payload.Version != Version {
		return fmt.Errorf("schema validation failed: version must be %s", Version)
	}
	if payload.ExportedAt == "" {
		return fmt.Errorf("schema validation failed: exported_at is required")
	}
	if payload.Source.Platform == "" || payload.Source.DeviceID == "" || payload.Source.AppVersion == "" {
		return fmt.Errorf("schema validation failed: source.platform, source.device_id, and source.app_version are required")
	}
	for _, identity := range payload.Identities {
		if identity.ID == "" || identity.Type == "" || identity.DisplayName == "" {
			return fmt.Errorf("schema validation failed: identity id/type/display_name required")
		}
		if identity.Phones == nil || identity.Emails == nil || identity.Labels == nil {
			return fmt.Errorf("schema validation failed: identity phones/emails/labels must be present")
		}
		switch identity.Type {
		case "person", "device", "account":
		default:
			return fmt.Errorf("schema validation failed: unsupported identity type %q", identity.Type)
		}
	}
	for _, event := range payload.Events {
		if event.ID == "" || event.Timestamp == "" || event.Type == "" {
			return fmt.Errorf("schema validation failed: event id/type/timestamp required")
		}
		if len(event.Participants) == 0 {
			return fmt.Errorf("schema validation failed: participants must contain at least one item")
		}
		for _, participant := range event.Participants {
			if participant == "" {
				return fmt.Errorf("schema validation failed: participant ids must be non-empty")
			}
		}
		switch event.Type {
		case "sms":
			if event.Direction != "inbound" && event.Direction != "outbound" {
				return fmt.Errorf("schema validation failed: sms direction invalid")
			}
			if _, ok := event.Content["text"]; !ok {
				return fmt.Errorf("schema validation failed: sms content.text required")
			}
		case "call":
			if event.Direction != "inbound" && event.Direction != "outbound" && event.Direction != "missed" {
				return fmt.Errorf("schema validation failed: call direction invalid")
			}
			if _, ok := event.Content["duration_sec"]; !ok {
				return fmt.Errorf("schema validation failed: call content.duration_sec required")
			}
		case "voice":
			if event.Direction != "inbound" && event.Direction != "outbound" {
				return fmt.Errorf("schema validation failed: voice direction invalid")
			}
			if _, ok := event.Content["file"]; !ok {
				return fmt.Errorf("schema validation failed: voice content.file required")
			}
		case "contact_snapshot":
			if event.Direction != "system" {
				return fmt.Errorf("schema validation failed: contact_snapshot direction must be system")
			}
			if _, ok := event.Content["identity_id"]; !ok {
				return fmt.Errorf("schema validation failed: contact_snapshot content.identity_id required")
			}
		default:
			return fmt.Errorf("schema validation failed: unsupported event type %q", event.Type)
		}
		for _, relation := range event.Relations {
			switch relation.Type {
			case "same_thread", "reply_to", "derived_from", "references_identity":
			default:
				return fmt.Errorf("schema validation failed: unsupported relation type %q", relation.Type)
			}
		}
	}
	// TODO: additionalProperties parity still relies on the JSON Schema files until a full schema engine is introduced.
	return nil
}

func (v *Validator) ValidateFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return v.ValidateBytes(data)
}
