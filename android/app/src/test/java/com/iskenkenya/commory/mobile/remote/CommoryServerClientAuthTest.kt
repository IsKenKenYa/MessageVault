package com.iskenkenya.commory.mobile.remote

import android.content.Context
import androidx.test.core.app.ApplicationProvider
import com.iskenkenya.commory.mobile.runtime.AppEnvironmentManager
import com.iskenkenya.commory.mobile.runtime.AuthSession
import kotlinx.coroutines.runBlocking
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import okio.Buffer
import org.junit.After
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config
import kotlin.test.assertEquals
import kotlin.test.assertTrue

@RunWith(RobolectricTestRunner::class)
@Config(manifest = Config.NONE)
class CommoryServerClientAuthTest {
    private lateinit var context: Context
    private lateinit var environmentManager: AppEnvironmentManager
    private lateinit var server: MockWebServer
    private lateinit var client: CommoryServerClient

    @Before
    fun setUp() = runBlocking {
        context = ApplicationProvider.getApplicationContext()
        context.getSharedPreferences("commory_refresh_cookies", Context.MODE_PRIVATE).edit().clear().commit()
        environmentManager = AppEnvironmentManager(context)
        environmentManager.clearSession()
        server = MockWebServer()
        server.start()
        environmentManager.updateServerUrl(server.url("/").toString())
        client = CommoryServerClient(context, environmentManager)
    }

    @After
    fun tearDown() {
        server.shutdown()
        context.getSharedPreferences("commory_refresh_cookies", Context.MODE_PRIVATE).edit().clear().commit()
    }

    @Test
    fun refreshPersistedSessionRetriesOnceOnConflict() = runBlocking {
        val loginToken = fakeToken("user-1", "session-1", 3600L)
        server.enqueue(
            envelopeResponse(
                status = 200,
                body = """{"code":200,"msg":"ok","data":{"user":{"id":"user-1","userName":"alice","email":"alice@example.com","roles":["R_USER"]},"token":"$loginToken"}}""",
                setCookie = "commory_refresh_token=rt1; Path=/; HttpOnly; SameSite=Strict"
            )
        )

        val loginSession = client.login(server.url("/").toString(), "alice", "secret").getOrThrow()
        environmentManager.updateSession(loginSession)

        val refreshedToken = fakeToken("user-1", "session-1", 7200L)
        server.enqueue(envelopeResponse(409, """{"code":409,"msg":"ERR_REFRESH_TOKEN_RETRY","data":null}"""))
        server.enqueue(
            envelopeResponse(
                status = 200,
                body = """{"code":200,"msg":"ok","data":{"token":"$refreshedToken"}}""",
                setCookie = "commory_refresh_token=rt2; Path=/; HttpOnly; SameSite=Strict"
            )
        )
        server.enqueue(
            envelopeResponse(
                status = 200,
                body = """{"code":200,"msg":"ok","data":{"id":"user-1","userName":"alice","email":"alice@example.com","roles":["R_USER"]}}"""
            )
        )

        val refreshed = client.refreshPersistedSession().getOrThrow()

        assertEquals("user-1", refreshed.userId)
        assertEquals(refreshedToken, refreshed.accessToken)
        assertTrue(environmentManager.currentSnapshot().authSession.accessToken == refreshedToken)

        val loginRequest = server.takeRequest()
        assertEquals("/api/auth/login", loginRequest.path)

        val firstRefresh = server.takeRequest()
        assertEquals("/api/auth/refresh", firstRefresh.path)
        assertTrue(firstRefresh.getHeader("Cookie").orEmpty().contains("commory_refresh_token=rt1"))

        val secondRefresh = server.takeRequest()
        assertEquals("/api/auth/refresh", secondRefresh.path)
        assertTrue(secondRefresh.getHeader("Cookie").orEmpty().contains("commory_refresh_token=rt1"))

        val userInfo = server.takeRequest()
        assertEquals("/api/user/info", userInfo.path)
        assertEquals("Bearer $refreshedToken", userInfo.getHeader("Authorization"))
    }

    private fun envelopeResponse(status: Int, body: String, setCookie: String? = null): MockResponse {
        return MockResponse()
            .setResponseCode(status)
            .setHeader("Content-Type", "application/json")
            .apply {
                if (setCookie != null) {
                    setHeader("Set-Cookie", setCookie)
                }
            }
            .setBody(Buffer().writeUtf8(body))
    }

    private fun fakeToken(userId: String, sessionId: String, expiresInSeconds: Long): String {
        val header = """{"alg":"HS256","typ":"JWT"}""".base64Url()
        val payload = """{"sub":"$userId","sid":"$sessionId","exp":${(System.currentTimeMillis() / 1000L) + expiresInSeconds}}""".base64Url()
        return "$header.$payload.sig"
    }

    private fun String.base64Url(): String {
        return android.util.Base64.encodeToString(
            toByteArray(),
            android.util.Base64.URL_SAFE or android.util.Base64.NO_PADDING or android.util.Base64.NO_WRAP
        )
    }
}
