package msglayer

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestExamplesValidate(t *testing.T) {
	root := projectPath("msglayer", "schema", "v0.1", "root.schema.json")
	validator, err := NewValidator(root)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"export.minimal.json", "export.sms-call.json", "export.voice.json"} {
		if err := validator.ValidateFile(projectPath("msglayer", "examples", name)); err != nil {
			t.Fatalf("%s should validate: %v", name, err)
		}
	}
}

func TestValidationFailsForUnknownEventType(t *testing.T) {
	root := projectPath("msglayer", "schema", "v0.1", "root.schema.json")
	validator, err := NewValidator(root)
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte(`{
	  "version": "msglayer/v0.1",
	  "exported_at": "2026-01-01T12:00:00Z",
	  "source": {"platform":"android","device_id":"x","app_version":"0.1"},
	  "identities": [],
	  "events": [{"id":"e1","type":"fax","timestamp":"2026-01-01T12:00:00Z","direction":"inbound","participants":["self/x"],"content":{},"meta":{},"relations":[]}]
	}`)
	if err := validator.ValidateBytes(payload); err == nil {
		t.Fatal("expected schema validation to fail")
	}
}

func TestValidationFailsForEmptyParticipant(t *testing.T) {
	root := projectPath("msglayer", "schema", "v0.1", "root.schema.json")
	validator, err := NewValidator(root)
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte(`{
	  "version": "msglayer/v0.1",
	  "exported_at": "2026-01-01T12:00:00Z",
	  "source": {"platform":"android","device_id":"x","app_version":"0.1"},
	  "identities": [{"id":"self/x","type":"device","display_name":"x","phones":[],"emails":[],"labels":["self"],"meta":{}}],
	  "events": [{"id":"e1","type":"sms","timestamp":"2026-01-01T12:00:00Z","direction":"inbound","participants":[""],"content":{"text":"hi","attachments":[]},"meta":{},"relations":[]}]
	}`)
	if err := validator.ValidateBytes(payload); err == nil {
		t.Fatal("expected participant validation to fail")
	}
}

func TestValidationFailsForVoiceDirection(t *testing.T) {
	root := projectPath("msglayer", "schema", "v0.1", "root.schema.json")
	validator, err := NewValidator(root)
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte(`{
	  "version": "msglayer/v0.1",
	  "exported_at": "2026-01-01T12:00:00Z",
	  "source": {"platform":"android","device_id":"x","app_version":"0.1"},
	  "identities": [{"id":"self/x","type":"device","display_name":"x","phones":[],"emails":[],"labels":["self"],"meta":{}}],
	  "events": [{"id":"voice_1","type":"voice","timestamp":"2026-01-01T12:00:00Z","direction":"system","participants":["self/x"],"content":{"file":"file://voice/001.mp3","transcript":"hi","summary":"hi"},"meta":{},"relations":[]}]
	}`)
	if err := validator.ValidateBytes(payload); err == nil {
		t.Fatal("expected voice direction validation to fail")
	}
}

func TestValidationFailsForRelationType(t *testing.T) {
	root := projectPath("msglayer", "schema", "v0.1", "root.schema.json")
	validator, err := NewValidator(root)
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte(`{
	  "version": "msglayer/v0.1",
	  "exported_at": "2026-01-01T12:00:00Z",
	  "source": {"platform":"android","device_id":"x","app_version":"0.1"},
	  "identities": [{"id":"self/x","type":"device","display_name":"x","phones":[],"emails":[],"labels":["self"],"meta":{}}],
	  "events": [{"id":"e1","type":"sms","timestamp":"2026-01-01T12:00:00Z","direction":"inbound","participants":["self/x"],"content":{"text":"hi","attachments":[]},"meta":{},"relations":[{"type":"bad_relation","target":"x"}]}]
	}`)
	if err := validator.ValidateBytes(payload); err == nil {
		t.Fatal("expected relation validation to fail")
	}
}

func TestValidationFailsForUnknownEventField(t *testing.T) {
	root := projectPath("msglayer", "schema", "v0.1", "root.schema.json")
	validator, err := NewValidator(root)
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte(`{
	  "version": "msglayer/v0.1",
	  "exported_at": "2026-01-01T12:00:00Z",
	  "source": {"platform":"android","device_id":"x","app_version":"0.1"},
	  "identities": [{"id":"self/x","type":"device","display_name":"x","phones":[],"emails":[],"labels":["self"],"meta":{}}],
	  "events": [{"id":"e1","type":"sms","timestamp":"2026-01-01T12:00:00Z","direction":"inbound","participants":["self/x"],"content":{"text":"hi","attachments":[]},"meta":{},"relations":[],"unknown_field":"oops"}]
	}`)
	if err := validator.ValidateBytes(payload); err == nil {
		t.Fatal("expected validation to fail for unknown event field")
	}
}

func TestValidationFailsForUnknownIdentityField(t *testing.T) {
	root := projectPath("msglayer", "schema", "v0.1", "root.schema.json")
	validator, err := NewValidator(root)
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte(`{
	  "version": "msglayer/v0.1",
	  "exported_at": "2026-01-01T12:00:00Z",
	  "source": {"platform":"android","device_id":"x","app_version":"0.1"},
	  "identities": [{"id":"self/x","type":"device","display_name":"x","phones":[],"emails":[],"labels":["self"],"meta":{},"extra_field":"oops"}],
	  "events": [{"id":"e1","type":"sms","timestamp":"2026-01-01T12:00:00Z","direction":"inbound","participants":["self/x"],"content":{"text":"hi","attachments":[]},"meta":{},"relations":[]}]
	}`)
	if err := validator.ValidateBytes(payload); err == nil {
		t.Fatal("expected validation to fail for unknown identity field")
	}
}

func TestValidationFailsForUnknownContentKey(t *testing.T) {
	root := projectPath("msglayer", "schema", "v0.1", "root.schema.json")
	validator, err := NewValidator(root)
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte(`{
	  "version": "msglayer/v0.1",
	  "exported_at": "2026-01-01T12:00:00Z",
	  "source": {"platform":"android","device_id":"x","app_version":"0.1"},
	  "identities": [{"id":"self/x","type":"device","display_name":"x","phones":[],"emails":[],"labels":["self"],"meta":{}}],
	  "events": [{"id":"e1","type":"sms","timestamp":"2026-01-01T12:00:00Z","direction":"inbound","participants":["self/x"],"content":{"text":"hi","attachments":[],"unknown_key":"oops"},"meta":{},"relations":[]}]
	}`)
	if err := validator.ValidateBytes(payload); err == nil {
		t.Fatal("expected validation to fail for unknown content key")
	}
}

func TestValidationPassesForAllKnownFields(t *testing.T) {
	root := projectPath("msglayer", "schema", "v0.1", "root.schema.json")
	validator, err := NewValidator(root)
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte(`{
	  "version": "msglayer/v0.1",
	  "exported_at": "2026-01-01T12:00:00Z",
	  "source": {"platform":"android","device_id":"x","app_version":"0.1"},
	  "identities": [{"id":"self/x","type":"device","display_name":"x","phones":[],"emails":[],"labels":["self"],"meta":{}}],
	  "events": [{"id":"e1","type":"sms","timestamp":"2026-01-01T12:00:00Z","direction":"inbound","participants":["self/x"],"content":{"text":"hi","attachments":[]},"meta":{},"relations":[]}]
	}`)
	if err := validator.ValidateBytes(payload); err != nil {
		t.Fatalf("expected validation to pass for valid event with all known fields: %v", err)
	}
}

func projectPath(parts ...string) string {
	_, file, _, _ := runtime.Caller(0)
	base := filepath.Join(filepath.Dir(file), "..", "..", "..")
	all := append([]string{base}, parts...)
	return filepath.Join(all...)
}
