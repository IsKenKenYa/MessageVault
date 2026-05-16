package com.iskenkenya.commory.mobile.runtime

import kotlin.test.Test
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class RuntimeModePolicyTest {
    @Test
    fun serverModeRequiresAuthUntilSessionExists() {
        val anonymous = AppEnvironment(
            mode = RuntimeMode.COMMORY_SERVER,
            modeSelected = true
        )
        assertTrue(RuntimeModePolicy.requiresServerAuth(anonymous))

        val authenticated = anonymous.copy(
            authSession = AuthSession(accessToken = "access", userId = "user")
        )
        assertFalse(RuntimeModePolicy.requiresServerAuth(authenticated))
    }

    @Test
    fun expiredAccessTokenDoesNotAuthenticateSession() {
        val expired = AuthSession(
            accessToken = "access",
            accessTokenExpiresAtEpochSeconds = (System.currentTimeMillis() / 1000L) - 60L
        )

        assertFalse(expired.isAuthenticated)
    }

    @Test
    fun uploadRequiresServerModeSyncAndAuth() {
        val local = AppEnvironment(
            mode = RuntimeMode.LOCAL_ONLY,
            modeSelected = true,
            syncOnBackup = true,
            authSession = AuthSession(accessToken = "access")
        )
        assertFalse(RuntimeModePolicy.canUploadBackup(local))

        val serverNoSync = local.copy(mode = RuntimeMode.COMMORY_SERVER, syncOnBackup = false)
        assertFalse(RuntimeModePolicy.canUploadBackup(serverNoSync))

        val serverSync = serverNoSync.copy(syncOnBackup = true)
        assertTrue(RuntimeModePolicy.canUploadBackup(serverSync))
    }
}
