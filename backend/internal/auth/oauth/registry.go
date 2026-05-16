package oauth

import "sync"

// Registry 管理所有已注册的 OAuth Provider。
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewRegistry 创建一个空的 Provider 注册表。
func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

// Register 注册一个 OAuth Provider。
func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
}

// Get 按名称获取 Provider，不存在返回 nil。
func (r *Registry) Get(name string) Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.providers[name]
}

// ListEnabled 返回所有已启用的 Provider 列表。
func (r *Registry) ListEnabled() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []Provider
	for _, p := range r.providers {
		if p.IsEnabled() {
			result = append(result, p)
		}
	}
	return result
}

// ListAll 返回所有已注册的 Provider（包括未启用的）。
func (r *Registry) ListAll() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Provider, 0, len(r.providers))
	for _, p := range r.providers {
		result = append(result, p)
	}
	return result
}
