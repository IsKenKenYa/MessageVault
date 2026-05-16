package com.iskenkenya.commory.mobile.remote

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull
import okhttp3.Cookie
import okhttp3.HttpUrl.Companion.toHttpUrl

class CommoryRefreshCookieJarTest {
    @Test
    fun savesLoadsReplacesAndClearsPerOrigin() {
        val store = InMemoryRefreshCookieStore()
        val jar = CommoryRefreshCookieJar(store)
        val origin = "https://commory.example.com/api".toHttpUrl()

        jar.saveFromResponse(origin, listOf(cookie(origin.host, "rt1", System.currentTimeMillis() + 60_000L)))
        assertEquals("rt1", jar.loadForRequest(origin).single().value)

        jar.saveFromResponse(origin, listOf(cookie(origin.host, "rt2", System.currentTimeMillis() + 60_000L)))
        assertEquals("rt2", jar.loadForRequest(origin).single().value)

        jar.clear("https://commory.example.com/")
        assertEquals(emptyList(), jar.loadForRequest(origin))
        assertNull(store.get(CommoryRefreshCookieJar.originKey(origin)))
    }

    @Test
    fun expiredCookieIsDiscarded() {
        val jar = CommoryRefreshCookieJar(InMemoryRefreshCookieStore())
        val origin = "https://commory.example.com/api".toHttpUrl()

        jar.saveFromResponse(origin, listOf(cookie(origin.host, "expired", System.currentTimeMillis() - 1L)))

        assertEquals(emptyList(), jar.loadForRequest(origin))
    }

    private fun cookie(host: String, value: String, expiresAt: Long): Cookie {
        return Cookie.Builder()
            .name("commory_refresh_token")
            .value(value)
            .hostOnlyDomain(host)
            .path("/")
            .expiresAt(expiresAt)
            .httpOnly()
            .secure()
            .build()
    }

    private class InMemoryRefreshCookieStore : RefreshCookieStore {
        private val values = mutableMapOf<String, StoredRefreshCookie>()

        override fun get(origin: String): StoredRefreshCookie? = values[origin]

        override fun put(origin: String, cookie: StoredRefreshCookie) {
            values[origin] = cookie
        }

        override fun remove(origin: String) {
            values.remove(origin)
        }
    }
}
