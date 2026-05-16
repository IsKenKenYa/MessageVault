package com.iskenkenya.commory.mobile.remote

import android.content.Context
import android.util.Base64
import android.os.Build
import com.iskenkenya.commory.mobile.runtime.AppEnvironmentManager
import com.iskenkenya.commory.mobile.runtime.AuthSession
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.withContext
import okhttp3.MediaType
import okhttp3.OkHttpClient
import okhttp3.RequestBody
import okhttp3.Response
import org.json.JSONObject
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import java.io.File
import java.io.IOException
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.TimeUnit

class CommoryServerClient(
    private val context: Context,
    private val environmentManager: AppEnvironmentManager
) {
    private val serviceCache = ConcurrentHashMap<String, CommoryApiService>()
    private val refreshServiceCache = ConcurrentHashMap<String, CommoryApiService>()
    private val deviceName = buildDeviceName()

    private val refreshClient = baseClientBuilder().build()

    private val client = baseClientBuilder()
        .authenticator { _, response ->
            if (responseCount(response) > 1 || response.request().url().encodedPath().contains("/api/auth/")) {
                return@authenticator null
            }
            val refreshed = runBlocking { refreshPersistedSession().getOrNull() } ?: return@authenticator null
            val accessToken = refreshed.accessToken ?: return@authenticator null
            response.request().newBuilder()
                .header("Authorization", bearer(accessToken))
                .build()
        }
        .build()

    private fun baseClientBuilder(): OkHttpClient.Builder = OkHttpClient.Builder()
        .connectTimeout(15, TimeUnit.SECONDS)
        .readTimeout(30, TimeUnit.SECONDS)
        .writeTimeout(30, TimeUnit.SECONDS)
        .addInterceptor { chain ->
            val request = chain.request().newBuilder()
                .header("X-Commory-Device", deviceName)
                .build()
            chain.proceed(request)
        }

    private fun service(baseUrl: String): CommoryApiService {
        val normalized = normalizeBaseUrl(baseUrl)
        return serviceCache.getOrPut(normalized) { createService(normalized, client) }
    }

    private fun refreshService(baseUrl: String): CommoryApiService {
        val normalized = normalizeBaseUrl(baseUrl)
        return refreshServiceCache.getOrPut(normalized) { createService(normalized, refreshClient) }
    }

    private fun createService(baseUrl: String, okHttpClient: OkHttpClient): CommoryApiService {
        return Retrofit.Builder()
            .baseUrl(baseUrl)
            .client(okHttpClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
            .create(CommoryApiService::class.java)
    }

    suspend fun checkSetup(baseUrl: String): Result<SetupStatusDto> = wrapCall {
        val response = service(baseUrl).getSetupStatus()
        require(response.isSuccessful) { response.errorBody()?.string() ?: "setup request failed" }
        response.body()?.data ?: error(response.body()?.msg ?: "missing setup payload")
    }

    suspend fun register(baseUrl: String, userName: String, email: String, password: String): Result<AuthSession> = wrapCall {
        val response = service(baseUrl).register(RegisterRequestDto(userName, email, password))
        require(response.isSuccessful) { response.body()?.msg ?: response.errorBody()?.string() ?: "register failed" }
        response.body()?.data?.toSession() ?: error(response.body()?.msg ?: "missing register payload")
    }

    suspend fun login(baseUrl: String, userName: String, password: String): Result<AuthSession> = wrapCall {
        val response = service(baseUrl).login(LoginRequestDto(userName, password))
        require(response.isSuccessful) { response.body()?.msg ?: response.errorBody()?.string() ?: "login failed" }
        response.body()?.data?.toSession() ?: error(response.body()?.msg ?: "missing login payload")
    }

    suspend fun refresh(baseUrl: String, refreshToken: String, currentSession: AuthSession): Result<AuthSession> = wrapCall {
        val response = refreshService(baseUrl).refresh(RefreshRequestDto(refreshToken))
        require(response.isSuccessful) { response.body()?.msg ?: response.errorBody()?.string() ?: "refresh failed" }
        val pair = response.body()?.data ?: error(response.body()?.msg ?: "missing refresh payload")
        val accessToken = pair.accessToken ?: pair.token ?: error("missing access token")
        currentSession.copy(
            accessToken = accessToken,
            refreshToken = pair.refreshToken,
            accessTokenExpiresAtEpochSeconds = accessTokenExpiresAt(accessToken),
            sessionId = accessTokenSessionId(accessToken) ?: currentSession.sessionId,
            deviceName = currentSession.deviceName ?: deviceName
        )
    }

    suspend fun logoutPersistedSession(): Result<Unit> = wrapCall {
        val environment = environmentManager.currentSnapshot()
        val refreshToken = environment.authSession.refreshToken
        if (!refreshToken.isNullOrBlank()) {
            val response = refreshService(environment.serverUrl).logout(LogoutRequestDto(refreshToken))
            require(response.isSuccessful) { response.body()?.msg ?: response.errorBody()?.string() ?: "logout failed" }
        }
        environmentManager.clearSession()
    }

    suspend fun userInfo(baseUrl: String, accessToken: String): Result<UserDto> = wrapCall {
        val response = service(baseUrl).getUserInfo(bearer(accessToken))
        require(response.isSuccessful) { response.body()?.msg ?: response.errorBody()?.string() ?: "user info failed" }
        response.body()?.data ?: error(response.body()?.msg ?: "missing user payload")
    }

    suspend fun listImports(baseUrl: String, accessToken: String): Result<List<ImportSummaryDto>> = wrapCall {
        val response = service(baseUrl).listImports(bearer(accessToken))
        require(response.isSuccessful) { response.body()?.msg ?: response.errorBody()?.string() ?: "list imports failed" }
        response.body()?.data ?: emptyList()
    }

    suspend fun uploadImport(baseUrl: String, accessToken: String, file: File): Result<ImportUploadResponseDto> = wrapCall {
        val body = withContext(Dispatchers.IO) {
            RequestBody.create(MediaType.parse("application/json"), file.readBytes())
        }
        val response = service(baseUrl).uploadImport(bearer(accessToken), body)
        require(response.isSuccessful) { response.body()?.msg ?: response.errorBody()?.string() ?: "upload failed" }
        response.body()?.data ?: error(response.body()?.msg ?: "missing upload payload")
    }

    suspend fun exportImport(baseUrl: String, accessToken: String, importId: String): Result<File> = wrapCall {
        val response = service(baseUrl).exportImport(bearer(accessToken), importId)
        require(response.isSuccessful) { response.errorBody()?.string() ?: "export failed" }
        val body = response.body() ?: error("missing export body")
        val outDir = File(context.cacheDir, "remote-imports").apply { mkdirs() }
        val outFile = File(outDir, "$importId.json")
        withContext(Dispatchers.IO) {
            body.byteStream().use { input ->
                outFile.outputStream().use { output -> input.copyTo(output) }
            }
        }
        outFile
    }

    suspend fun refreshPersistedSession(): Result<AuthSession> = wrapCall {
        val environment = environmentManager.currentSnapshot()
        val refreshToken = environment.authSession.refreshToken ?: error("missing refresh token")
        val updated = refresh(environment.serverUrl, refreshToken, environment.authSession).getOrElse { throwable ->
            if (throwable.isTerminalAuthFailure()) {
                environmentManager.clearSession()
            }
            throw throwable
        }
        val userResponse = refreshService(environment.serverUrl).getUserInfo(bearer(updated.accessToken ?: ""))
        require(userResponse.isSuccessful) { userResponse.body()?.msg ?: userResponse.errorBody()?.string() ?: "user info failed" }
        val user = userResponse.body()?.data ?: error(userResponse.body()?.msg ?: "missing user payload")
        val session = updated.copy(
            userId = user.id,
            userName = user.userName ?: user.user_name,
            email = user.email,
            deviceName = updated.deviceName ?: deviceName
        )
        environmentManager.updateSession(session)
        session
    }

    private fun normalizeBaseUrl(baseUrl: String): String {
        val trimmed = baseUrl.trim()
        val withScheme = if (trimmed.startsWith("http://") || trimmed.startsWith("https://")) trimmed else "http://$trimmed"
        return if (withScheme.endsWith("/")) withScheme else "$withScheme/"
    }

    private fun bearer(token: String): String = "Bearer $token"

    private suspend fun <T> wrapCall(block: suspend () -> T): Result<T> {
        return runCatching { block() }
            .fold(
                onSuccess = { Result.success(it) },
                onFailure = { throwable ->
                    val failure = when (throwable) {
                        is NetworkException -> throwable
                        is IOException -> throwable.asNetworkException()
                        is IllegalArgumentException -> NetworkException(NetworkError.Validation(throwable.message ?: "invalid request"))
                        is IllegalStateException -> NetworkException(NetworkError.Validation(throwable.message ?: "invalid response"))
                        else -> throwable.asNetworkException()
                    }
                    Result.failure(failure)
                }
            )
    }

    private fun AuthResponseDto.toSession(): AuthSession {
        return AuthSession(
            accessToken = token,
            refreshToken = refreshToken,
            accessTokenExpiresAtEpochSeconds = accessTokenExpiresAt(token),
            sessionId = accessTokenSessionId(token),
            deviceName = deviceName,
            userId = user.id,
            userName = user.userName ?: user.user_name,
            email = user.email
        )
    }

    private fun accessTokenExpiresAt(token: String?): Long? {
        if (token.isNullOrBlank()) return null
        return runCatching {
            val payload = token.split(".").getOrNull(1) ?: return null
            val decoded = Base64.decode(payload, Base64.URL_SAFE or Base64.NO_PADDING or Base64.NO_WRAP)
            JSONObject(String(decoded)).optLong("exp").takeIf { it > 0L }
        }.getOrNull()
    }

    private fun accessTokenSessionId(token: String?): String? {
        if (token.isNullOrBlank()) return null
        return runCatching {
            val payload = token.split(".").getOrNull(1) ?: return null
            val decoded = Base64.decode(payload, Base64.URL_SAFE or Base64.NO_PADDING or Base64.NO_WRAP)
            JSONObject(String(decoded)).optString("sid").takeIf { it.isNotBlank() }
        }.getOrNull()
    }

    private fun responseCount(response: Response): Int {
        var count = 1
        var prior = response.priorResponse()
        while (prior != null) {
            count++
            prior = prior.priorResponse()
        }
        return count
    }

    private fun buildDeviceName(): String {
        val manufacturer = Build.MANUFACTURER?.trim().orEmpty()
        val model = Build.MODEL?.trim().orEmpty()
        return listOf(manufacturer, model)
            .filter { it.isNotBlank() }
            .distinct()
            .joinToString(" ")
            .ifBlank { "Android Device" }
    }

    private fun Throwable.isTerminalAuthFailure(): Boolean {
        val network = this as? NetworkException ?: return false
        return network.error is NetworkError.Unauthorized
    }
}
