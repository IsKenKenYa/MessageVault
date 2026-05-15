package agent

import (
	"context"
	"errors"
	"fmt"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Message struct {
	Role    Role
	Content string
}

type ToolSpec struct {
	Name        string
	Description string
	InputSchema string
}

type Request struct {
	SessionID string
	Messages  []Message
	Tools     []ToolSpec
}

type Response struct {
	Message Message
}

type Provider interface {
	Complete(context.Context, Request) (Response, error)
}

type ToolRegistry interface {
	ListTools(context.Context, string) ([]ToolSpec, error)
}

type PermissionDecision string

const (
	PermissionAllow PermissionDecision = "allow"
	PermissionDeny  PermissionDecision = "deny"
)

type PermissionRequest struct {
	SessionID string
	ToolName  string
}

type PermissionBroker interface {
	Decide(context.Context, PermissionRequest) (PermissionDecision, error)
}

type SessionStore interface {
	Append(context.Context, string, []Message) error
}

type Runtime struct {
	provider    Provider
	tools       ToolRegistry
	permissions PermissionBroker
	sessions    SessionStore
}

func NewRuntime(provider Provider, tools ToolRegistry, permissions PermissionBroker, sessions SessionStore) (*Runtime, error) {
	if provider == nil {
		return nil, errors.New("agent provider is required")
	}
	if tools == nil {
		return nil, errors.New("agent tool registry is required")
	}
	if permissions == nil {
		return nil, errors.New("agent permission broker is required")
	}
	if sessions == nil {
		return nil, errors.New("agent session store is required")
	}
	return &Runtime{
		provider:    provider,
		tools:       tools,
		permissions: permissions,
		sessions:    sessions,
	}, nil
}

func (r *Runtime) Run(ctx context.Context, sessionID string, input Message) (Message, error) {
	if sessionID == "" {
		return Message{}, errors.New("agent session id is required")
	}
	if input.Role == "" {
		return Message{}, errors.New("agent input role is required")
	}
	if err := r.sessions.Append(ctx, sessionID, []Message{input}); err != nil {
		return Message{}, fmt.Errorf("append user input: %w", err)
	}
	tools, err := r.tools.ListTools(ctx, sessionID)
	if err != nil {
		return Message{}, fmt.Errorf("list tools: %w", err)
	}
	for _, tool := range tools {
		decision, err := r.permissions.Decide(ctx, PermissionRequest{SessionID: sessionID, ToolName: tool.Name})
		if err != nil {
			return Message{}, fmt.Errorf("decide permission for %s: %w", tool.Name, err)
		}
		if decision != PermissionAllow {
			return Message{}, fmt.Errorf("tool %s permission denied", tool.Name)
		}
	}
	response, err := r.provider.Complete(ctx, Request{
		SessionID: sessionID,
		Messages:  []Message{input},
		Tools:     tools,
	})
	if err != nil {
		return Message{}, fmt.Errorf("complete agent response: %w", err)
	}
	if response.Message.Role == "" {
		response.Message.Role = RoleAssistant
	}
	if err := r.sessions.Append(ctx, sessionID, []Message{response.Message}); err != nil {
		return Message{}, fmt.Errorf("append assistant response: %w", err)
	}
	return response.Message, nil
}
