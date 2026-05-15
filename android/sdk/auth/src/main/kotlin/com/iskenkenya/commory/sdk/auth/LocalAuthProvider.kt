package com.iskenkenya.commory.sdk.auth

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

class LocalAuthProvider(private val deviceId: String) : AuthProvider {

    private val _isAuthenticated = MutableStateFlow(true)
    override val isAuthenticated: StateFlow<Boolean> = _isAuthenticated.asStateFlow()

    private val localUser = UserInfo(
        userId = "local_$deviceId",
        displayName = "Local User"
    )

    private val _currentUser = MutableStateFlow<UserInfo?>(localUser)
    override val currentUser: StateFlow<UserInfo?> = _currentUser.asStateFlow()

    private var token: String = "local_token_$deviceId"

    override suspend fun login(credentials: AuthCredentials): AuthResult {
        return when (credentials) {
            is AuthCredentials.Local -> {
                _isAuthenticated.value = true
                _currentUser.value = localUser
                AuthResult.Success(token, localUser)
            }
            else -> AuthResult.Error("UNSUPPORTED", "LocalAuthProvider only supports local credentials")
        }
    }

    override suspend fun logout() {
        _isAuthenticated.value = true
        _currentUser.value = localUser
    }

    override suspend fun refreshToken(): AuthResult {
        return AuthResult.Success(token, localUser)
    }

    override fun getToken(): String = token
}
