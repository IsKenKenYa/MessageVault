package com.iskenkenya.commory.sdk.auth

interface ThirdPartyAuthProvider : AuthProvider {
    val providerId: String
    val providerName: String

    suspend fun initialize(config: Map<String, String>): Boolean
    suspend fun handleCallback(callbackData: Map<String, String>): AuthResult
    fun getAuthUrl(): String
}
