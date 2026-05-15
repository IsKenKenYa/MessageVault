package com.iskenkenya.commory.sdk.auth

import kotlinx.coroutines.flow.StateFlow

interface AuthProvider {
    val isAuthenticated: StateFlow<Boolean>
    val currentUser: StateFlow<UserInfo?>

    suspend fun login(credentials: AuthCredentials): AuthResult
    suspend fun logout()
    suspend fun refreshToken(): AuthResult
    fun getToken(): String?
}

data class UserInfo(
    val userId: String,
    val displayName: String,
    val email: String? = null,
    val avatarUrl: String? = null
)

sealed class AuthResult {
    data class Success(val token: String, val user: UserInfo) : AuthResult()
    data class Error(val code: String, val message: String) : AuthResult()
    data object Cancelled : AuthResult()
}

sealed class AuthCredentials {
    data class Local(val deviceId: String) : AuthCredentials()
    data class OAuth2(val provider: String, val accessToken: String) : AuthCredentials()
    data class Password(val username: String, val password: String) : AuthCredentials()
}
