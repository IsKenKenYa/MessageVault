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

func projectPath(parts ...string) string {
	_, file, _, _ := runtime.Caller(0)
	base := filepath.Join(filepath.Dir(file), "..", "..", "..")
	all := append([]string{base}, parts...)
	return filepath.Join(all...)
}
