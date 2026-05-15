package imken.messagevault.mobile.runtime

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.runBlocking

private val Context.appEnvironmentStore: DataStore<Preferences> by preferencesDataStore(name = "commory_app_environment")

class AppEnvironmentManager(context: Context) {
    private val appContext = context.applicationContext
    private val dataStore = appContext.appEnvironmentStore

    val environment: Flow<AppEnvironment> = dataStore.data.map { preferences ->
        AppEnvironment(
            mode = RuntimeMode.valueOf(
                preferences[KEY_RUNTIME_MODE] ?: RuntimeMode.LOCAL_ONLY.name
            ),
            modeSelected = preferences[KEY_MODE_SELECTED] ?: false,
            serverUrl = preferences[KEY_SERVER_URL] ?: DEFAULT_SERVER_URL,
            locale = AppLocaleOption.valueOf(
                preferences[KEY_LOCALE] ?: AppLocaleOption.SYSTEM.name
            ),
            syncOnBackup = preferences[KEY_SYNC_ON_BACKUP] ?: true,
            authSession = AuthSession(
                accessToken = preferences[KEY_ACCESS_TOKEN],
                refreshToken = preferences[KEY_REFRESH_TOKEN],
                userId = preferences[KEY_USER_ID],
                userName = preferences[KEY_USER_NAME],
                email = preferences[KEY_EMAIL]
            )
        )
    }

    fun currentSnapshot(): AppEnvironment = runBlocking { environment.first() }

    suspend fun selectMode(mode: RuntimeMode) {
        dataStore.edit { preferences ->
            preferences[KEY_RUNTIME_MODE] = mode.name
            preferences[KEY_MODE_SELECTED] = true
        }
    }

    suspend fun updateServerUrl(serverUrl: String) {
        dataStore.edit { preferences ->
            preferences[KEY_SERVER_URL] = normalizeServerUrl(serverUrl)
        }
    }

    suspend fun updateLocale(locale: AppLocaleOption) {
        dataStore.edit { preferences ->
            preferences[KEY_LOCALE] = locale.name
        }
    }

    suspend fun updateSyncOnBackup(enabled: Boolean) {
        dataStore.edit { preferences ->
            preferences[KEY_SYNC_ON_BACKUP] = enabled
        }
    }

    suspend fun updateSession(session: AuthSession) {
        dataStore.edit { preferences ->
            setOrRemove(preferences, KEY_ACCESS_TOKEN, session.accessToken)
            setOrRemove(preferences, KEY_REFRESH_TOKEN, session.refreshToken)
            setOrRemove(preferences, KEY_USER_ID, session.userId)
            setOrRemove(preferences, KEY_USER_NAME, session.userName)
            setOrRemove(preferences, KEY_EMAIL, session.email)
        }
    }

    suspend fun clearSession() {
        updateSession(AuthSession())
    }

    private fun normalizeServerUrl(input: String): String {
        val trimmed = input.trim()
        if (trimmed.isBlank()) {
            return DEFAULT_SERVER_URL
        }
        val withScheme = if (trimmed.startsWith("http://") || trimmed.startsWith("https://")) {
            trimmed
        } else {
            "http://$trimmed"
        }
        return if (withScheme.endsWith("/")) withScheme else "$withScheme/"
    }

    private fun setOrRemove(
        preferences: androidx.datastore.preferences.core.MutablePreferences,
        key: Preferences.Key<String>,
        value: String?
    ) {
        if (value.isNullOrBlank()) {
            preferences.remove(key)
        } else {
            preferences[key] = value
        }
    }

    private companion object {
        const val DEFAULT_SERVER_URL = "http://10.0.2.2:3000/"

        val KEY_RUNTIME_MODE = stringPreferencesKey("runtime_mode")
        val KEY_MODE_SELECTED = booleanPreferencesKey("mode_selected")
        val KEY_SERVER_URL = stringPreferencesKey("server_url")
        val KEY_LOCALE = stringPreferencesKey("locale")
        val KEY_SYNC_ON_BACKUP = booleanPreferencesKey("sync_on_backup")
        val KEY_ACCESS_TOKEN = stringPreferencesKey("access_token")
        val KEY_REFRESH_TOKEN = stringPreferencesKey("refresh_token")
        val KEY_USER_ID = stringPreferencesKey("user_id")
        val KEY_USER_NAME = stringPreferencesKey("user_name")
        val KEY_EMAIL = stringPreferencesKey("email")
    }
}
