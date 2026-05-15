package agent

import (
	"context"
	"testing"
)

func TestRuntimeRunPersistsInputAndResponse(t *testing.T) {
	store := &memorySessionStore{}
	runtime, err := NewRuntime(
		providerFunc(func(ctx context.Context, req Request) (Response, error) {
			if req.SessionID != "session-1" {
				t.Fatalf("unexpected session id: %s", req.SessionID)
			}
			if len(req.Tools) != 1 || req.Tools[0].Name != "read_file" {
				t.Fatalf("unexpected tools: %#v", req.Tools)
			}
			return Response{Message: Message{Content: "done"}}, nil
		}),
		toolRegistryFunc(func(ctx context.Context, sessionID string) ([]ToolSpec, error) {
			return []ToolSpec{{Name: "read_file", Description: "Read a file"}}, nil
		}),
		permissionBrokerFunc(func(ctx context.Context, req PermissionRequest) (PermissionDecision, error) {
			if req.ToolName != "read_file" {
				t.Fatalf("unexpected tool permission request: %s", req.ToolName)
			}
			return PermissionAllow, nil
		}),
		store,
	)
	if err != nil {
		t.Fatal(err)
	}

	message, err := runtime.Run(context.Background(), "session-1", Message{Role: RoleUser, Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if message.Role != RoleAssistant || message.Content != "done" {
		t.Fatalf("unexpected message: %#v", message)
	}
	if len(store.messages) != 2 {
		t.Fatalf("expected two persisted messages, got %d", len(store.messages))
	}
}

func TestRuntimeRunDeniesToolsBeforeProviderCall(t *testing.T) {
	called := false
	runtime, err := NewRuntime(
		providerFunc(func(ctx context.Context, req Request) (Response, error) {
			called = true
			return Response{}, nil
		}),
		toolRegistryFunc(func(ctx context.Context, sessionID string) ([]ToolSpec, error) {
			return []ToolSpec{{Name: "write_file"}}, nil
		}),
		permissionBrokerFunc(func(ctx context.Context, req PermissionRequest) (PermissionDecision, error) {
			return PermissionDeny, nil
		}),
		&memorySessionStore{},
	)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := runtime.Run(context.Background(), "session-1", Message{Role: RoleUser, Content: "hello"}); err == nil {
		t.Fatal("expected permission error")
	}
	if called {
		t.Fatal("provider was called after permission denial")
	}
}

type providerFunc func(context.Context, Request) (Response, error)

func (f providerFunc) Complete(ctx context.Context, req Request) (Response, error) {
	return f(ctx, req)
}

type toolRegistryFunc func(context.Context, string) ([]ToolSpec, error)

func (f toolRegistryFunc) ListTools(ctx context.Context, sessionID string) ([]ToolSpec, error) {
	return f(ctx, sessionID)
}

type permissionBrokerFunc func(context.Context, PermissionRequest) (PermissionDecision, error)

func (f permissionBrokerFunc) Decide(ctx context.Context, req PermissionRequest) (PermissionDecision, error) {
	return f(ctx, req)
}

type memorySessionStore struct {
	messages []Message
}

func (s *memorySessionStore) Append(ctx context.Context, sessionID string, messages []Message) error {
	s.messages = append(s.messages, messages...)
	return nil
}
