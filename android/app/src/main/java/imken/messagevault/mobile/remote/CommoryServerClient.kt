package imken.messagevault.mobile.remote

import android.content.Context
import imken.messagevault.mobile.runtime.AppEnvironmentManager
import imken.messagevault.mobile.runtime.AuthSession
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.MediaType
import okhttp3.OkHttpClient
import okhttp3.RequestBody
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import java.io.File
import java.util.concurrent.TimeUnit

class CommoryServerClient(
    private val context: Context,
    private val environmentManager: AppEnvironmentManager
) {
    private val client = OkHttpClient.Builder()
        .connectTimeout(15, TimeUnit.SECONDS)
        .readTimeout(30, TimeUnit.SECONDS)
        .writeTimeout(30, TimeUnit.SECONDS)
        .build()

    private fun service(baseUrl: String): CommoryApiService {
        return Retrofit.Builder()
            .baseUrl(normalizeBaseUrl(baseUrl))
            .client(client)
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
        val response = service(baseUrl).refresh(RefreshRequestDto(refreshToken))
        require(response.isSuccessful) { response.body()?.msg ?: response.errorBody()?.string() ?: "refresh failed" }
        val pair = response.body()?.data ?: error(response.body()?.msg ?: "missing refresh payload")
        currentSession.copy(accessToken = pair.accessToken, refreshToken = pair.refreshToken)
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

    suspend fun refreshPersistedSession(): Result<AuthSession> {
        val environment = environmentManager.currentSnapshot()
        val refreshToken = environment.authSession.refreshToken ?: return Result.failure(IllegalStateException("missing refresh token"))
        val updated = refresh(environment.serverUrl, refreshToken, environment.authSession).getOrThrow()
        val user = userInfo(environment.serverUrl, updated.accessToken ?: "").getOrThrow()
        val session = updated.copy(
            userId = user.id,
            userName = user.userName ?: user.user_name,
            email = user.email
        )
        environmentManager.updateSession(session)
        return Result.success(session)
    }

    private fun normalizeBaseUrl(baseUrl: String): String {
        val trimmed = baseUrl.trim()
        val withScheme = if (trimmed.startsWith("http://") || trimmed.startsWith("https://")) trimmed else "http://$trimmed"
        return if (withScheme.endsWith("/")) withScheme else "$withScheme/"
    }

    private fun bearer(token: String): String = "Bearer $token"

    private suspend fun <T> wrapCall(block: suspend () -> T): Result<T> {
        return runCatching { block() }
    }

    private fun AuthResponseDto.toSession(): AuthSession {
        return AuthSession(
            accessToken = token,
            refreshToken = refreshToken,
            userId = user.id,
            userName = user.userName ?: user.user_name,
            email = user.email
        )
    }
}
