package imken.messagevault.sdk.auth

class AuthManager private constructor(
    private val provider: AuthProvider
) : AuthProvider by provider {

    companion object {
        fun createLocal(deviceId: String): AuthManager {
            return AuthManager(LocalAuthProvider(deviceId))
        }

        fun createWithProvider(provider: AuthProvider): AuthManager {
            return AuthManager(provider)
        }
    }
}
