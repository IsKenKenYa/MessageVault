package com.iskenkenya.commory.mobile.runtime

enum class RuntimeMode {
    LOCAL_ONLY,
    COMMORY_SERVER
}

enum class AppLocaleOption {
    SYSTEM,
    ZH_CN,
    EN
}

data class AuthSession(
    val accessToken: String? = null,
    val refreshToken: String? = null,
    val accessTokenExpiresAtEpochSeconds: Long? = null,
    val sessionId: String? = null,
    val deviceName: String? = null,
    val userId: String? = null,
    val userName: String? = null,
    val email: String? = null
) {
    val isAuthenticated: Boolean
        get() {
            if (accessToken.isNullOrBlank()) return false
            val expiresAt = accessTokenExpiresAtEpochSeconds ?: return true
            return (System.currentTimeMillis() / 1000L) < expiresAt
        }
}

data class AppEnvironment(
    val mode: RuntimeMode = RuntimeMode.LOCAL_ONLY,
    val modeSelected: Boolean = false,
    val serverUrl: String = "http://10.0.2.2:3000/",
    val locale: AppLocaleOption = AppLocaleOption.SYSTEM,
    val syncOnBackup: Boolean = true,
    val authSession: AuthSession = AuthSession()
)
