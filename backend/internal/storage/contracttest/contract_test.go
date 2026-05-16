package contracttest

import (
	"github.com/IsKenKenYa/Commory/backend/internal/storage"
	"testing"
)

// ProviderFactory 创建一个临时的 Provider 实例。
type ProviderFactory func(t *testing.T) storage.Provider

// RunContractTests 对给定的 Provider 运行完整的存储契约测试套件。
func RunContractTests(t *testing.T, factory ProviderFactory) {
	t.Run("UserLifecycle", func(t *testing.T) { testUserLifecycle(t, factory) })
	t.Run("RefreshTokenRotation", func(t *testing.T) { testRefreshTokenRotation(t, factory) })
	t.Run("ImportAndQuery", func(t *testing.T) { testImportAndQuery(t, factory) })
	t.Run("IdentityCRUD", func(t *testing.T) { testIdentityCRUD(t, factory) })
	t.Run("SearchPagination", func(t *testing.T) { testSearchPagination(t, factory) })
	t.Run("SetupLifecycle", func(t *testing.T) { testSetupLifecycle(t, factory) })
	t.Run("SessionLifecycle", func(t *testing.T) { testSessionLifecycle(t, factory) })
	t.Run("AuditLog", func(t *testing.T) { testAuditLog(t, factory) })
	t.Run("PasskeyCredential", func(t *testing.T) { testPasskeyCredential(t, factory) })
	t.Run("ChallengeLifecycle", func(t *testing.T) { testChallengeLifecycle(t, factory) })
	t.Run("AuthMethod", func(t *testing.T) { testAuthMethod(t, factory) })
}
