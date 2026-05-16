package com.iskenkenya.commory.mobile.remote

import android.content.Context
import okhttp3.Cookie
import okhttp3.CookieJar
import okhttp3.HttpUrl
import org.json.JSONObject

private const val REFRESH_COOKIE_NAME = "commory_refresh_token"
private const val COOKIE_STORE_NAME = "commory_refresh_cookies"

class CommoryRefreshCookieJar(
    private val store: RefreshCookieStore
) : CookieJar {
    override fun saveFromResponse(url: HttpUrl, cookies: List<Cookie>) {
        cookies.filter { it.name() == REFRESH_COOKIE_NAME }.forEach { cookie ->
            val origin = originKey(url)
            if (cookie.value().isBlank() || cookie.expiresAt() <= System.currentTimeMillis()) {
                store.remove(origin)
            } else {
                store.put(
                    origin,
                    StoredRefreshCookie(
                        value = cookie.value(),
                        expiresAt = cookie.expiresAt(),
                        secure = cookie.secure(),
                        path = cookie.path()
                    )
                )
            }
        }
    }

    override fun loadForRequest(url: HttpUrl): List<Cookie> {
        val stored = store.get(originKey(url)) ?: return emptyList()
        if (stored.expiresAt <= System.currentTimeMillis()) {
            store.remove(originKey(url))
            return emptyList()
        }
        if (stored.secure && url.scheme() != "https") {
            return emptyList()
        }
        return listOf(
            Cookie.Builder()
                .name(REFRESH_COOKIE_NAME)
                .value(stored.value)
                .hostOnlyDomain(url.host())
                .path(stored.path.ifBlank { "/" })
                .expiresAt(stored.expiresAt)
                .apply {
                    if (stored.secure) secure()
                }
                .build()
        )
    }

    fun clear(baseUrl: String) {
        HttpUrl.parse(baseUrl)?.let { store.remove(originKey(it)) }
    }

    companion object {
        fun persistent(context: Context): CommoryRefreshCookieJar {
            return CommoryRefreshCookieJar(SharedPreferencesRefreshCookieStore(context))
        }

        internal fun originKey(url: HttpUrl): String = "${url.scheme()}://${url.host()}:${url.port()}"
    }
}

data class StoredRefreshCookie(
    val value: String,
    val expiresAt: Long,
    val secure: Boolean,
    val path: String
)

interface RefreshCookieStore {
    fun get(origin: String): StoredRefreshCookie?
    fun put(origin: String, cookie: StoredRefreshCookie)
    fun remove(origin: String)
}

class SharedPreferencesRefreshCookieStore(context: Context) : RefreshCookieStore {
    private val preferences = context.applicationContext.getSharedPreferences(COOKIE_STORE_NAME, Context.MODE_PRIVATE)

    override fun get(origin: String): StoredRefreshCookie? {
        val raw = preferences.getString(origin, null) ?: return null
        return runCatching {
            val json = JSONObject(raw)
            StoredRefreshCookie(
                value = json.getString("value"),
                expiresAt = json.getLong("expiresAt"),
                secure = json.optBoolean("secure", false),
                path = json.optString("path", "/")
            )
        }.getOrNull()
    }

    override fun put(origin: String, cookie: StoredRefreshCookie) {
        val raw = JSONObject()
            .put("value", cookie.value)
            .put("expiresAt", cookie.expiresAt)
            .put("secure", cookie.secure)
            .put("path", cookie.path)
            .toString()
        preferences.edit().putString(origin, raw).apply()
    }

    override fun remove(origin: String) {
        preferences.edit().remove(origin).apply()
    }
}
