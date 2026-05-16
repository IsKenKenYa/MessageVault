package oauth

import "context"

// Provider 定义 OAuth 身份提供者的接口。
// 具体实现（GitHub、Google、OIDC 等）在后续 Phase 中添加。
type Provider interface {
	// Name 返回 provider 的唯一标识，如 "github", "google", "oidc"
	Name() string
	// DisplayName 返回用户可读的名称，如 "GitHub", "Google"
	DisplayName() string
	// IsEnabled 检查该 provider 的配置是否完整
	IsEnabled() bool
	// ExchangeCode 用授权码换取 access token
	ExchangeCode(ctx context.Context, code string) (*Token, error)
	// GetUserInfo 用 access token 获取用户信息
	GetUserInfo(ctx context.Context, token *Token) (*UserInfo, error)
}

// Token 表示 OAuth token 响应
type Token struct {
	AccessToken  string
	TokenType    string
	RefreshToken string
	ExpiresIn    int64
	IDToken      string // OIDC
}

// UserInfo 表示从 OAuth provider 获取的用户信息
type UserInfo struct {
	ProviderUserID string            // provider 内的用户唯一标识
	Username       string            // 用户名
	DisplayName    string            // 显示名
	Email          string            // 邮箱
	AvatarURL      string            // 头像
	Raw            map[string]any    // provider 特有的原始信息
}
