package imken.messagevault.mobile.auth

import imken.messagevault.mobile.remote.CommoryServerClient
import imken.messagevault.mobile.runtime.AppEnvironmentManager
import imken.messagevault.mobile.runtime.AuthSession
import imken.messagevault.sdk.auth.AuthCredentials
import imken.messagevault.sdk.auth.AuthProvider
import imken.messagevault.sdk.auth.AuthResult
import imken.messagevault.sdk.auth.UserInfo
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

class CommoryServerAuthProvider(
    private val environmentManager: AppEnvironmentManager,
    private val serverClient: CommoryServerClient
) : AuthProvider {
    private val _isAuthenticated = MutableStateFlow(environmentManager.currentSnapshot().authSession.isAuthenticated)
    override val isAuthenticated: StateFlow<Boolean> = _isAuthenticated.asStateFlow()

    private val _currentUser = MutableStateFlow(environmentManager.currentSnapshot().authSession.toUserInfo())
    override val currentUser: StateFlow<UserInfo?> = _currentUser.asStateFlow()

    override suspend fun login(credentials: AuthCredentials): AuthResult {
        val password = credentials as? AuthCredentials.Password
            ?: return AuthResult.Error("UNSUPPORTED", "Unsupported credentials")
        val environment = environmentManager.currentSnapshot()
        val session = serverClient.login(
            baseUrl = environment.serverUrl,
            userName = password.username,
            password = password.password
        ).getOrElse { return AuthResult.Error("LOGIN_FAILED", it.message ?: "login failed") }
        environmentManager.updateSession(session)
        _isAuthenticated.value = true
        _currentUser.value = session.toUserInfo()
        return AuthResult.Success(session.accessToken.orEmpty(), session.toUserInfo()!!)
    }

    override suspend fun logout() {
        environmentManager.clearSession()
        _isAuthenticated.value = false
        _currentUser.value = null
    }

    override suspend fun refreshToken(): AuthResult {
        val refreshed = serverClient.refreshPersistedSession()
            .getOrElse { return AuthResult.Error("REFRESH_FAILED", it.message ?: "refresh failed") }
        _isAuthenticated.value = true
        _currentUser.value = refreshed.toUserInfo()
        return AuthResult.Success(refreshed.accessToken.orEmpty(), refreshed.toUserInfo()!!)
    }

    override fun getToken(): String? = environmentManager.currentSnapshot().authSession.accessToken

    private fun AuthSession.toUserInfo(): UserInfo? {
        val id = userId ?: return null
        return UserInfo(
            userId = id,
            displayName = userName ?: email ?: id,
            email = email
        )
    }
}
