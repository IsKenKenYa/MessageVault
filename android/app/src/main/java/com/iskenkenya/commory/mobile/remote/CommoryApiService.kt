package com.iskenkenya.commory.mobile.remote

import okhttp3.RequestBody
import okhttp3.ResponseBody
import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.Header
import retrofit2.http.POST
import retrofit2.http.Path
import retrofit2.http.Streaming

data class ApiEnvelope<T>(
    val code: Int,
    val msg: String,
    val data: T?
)

data class SetupStatusDto(
    val status: Boolean? = null,
    val initialized: Boolean? = null,
    val version: String? = null,
    val database_type: String? = null,
    val databaseType: String? = null,
    val root_init: Boolean? = null,
    val rootInit: Boolean? = null
)

data class UserDto(
    val id: String,
    val user_name: String? = null,
    val userName: String? = null,
    val email: String? = null,
    val roles: List<String> = emptyList()
)

data class TokenPairDto(
    val accessToken: String? = null,
    val token: String? = null,
    val refreshToken: String
)

data class AuthResponseDto(
    val user: UserDto,
    val token: String,
    val refreshToken: String
)

data class RegisterRequestDto(
    val userName: String,
    val email: String,
    val password: String
)

data class LoginRequestDto(
    val userName: String,
    val password: String
)

data class RefreshRequestDto(
    val refreshToken: String
)

data class LogoutRequestDto(
    val refreshToken: String
)

data class ImportSummaryDto(
    val id: String,
    val user_id: String? = null,
    val userId: String? = null,
    val schema_version: String? = null,
    val schemaVersion: String? = null,
    val imported_at: String? = null,
    val importedAt: String? = null,
    val source_path: String? = null,
    val sourcePath: String? = null,
    val event_count: Int? = null,
    val eventCount: Int? = null,
    val identity_count: Int? = null,
    val identityCount: Int? = null
)

data class ImportUploadResponseDto(
    val import_id: String? = null,
    val importId: String? = null,
    val msglayer_version: String? = null,
    val msglayerVersion: String? = null
)

interface CommoryApiService {
    @GET("api/setup")
    suspend fun getSetupStatus(): Response<ApiEnvelope<SetupStatusDto>>

    @POST("api/auth/register")
    suspend fun register(@Body request: RegisterRequestDto): Response<ApiEnvelope<AuthResponseDto>>

    @POST("api/auth/login")
    suspend fun login(@Body request: LoginRequestDto): Response<ApiEnvelope<AuthResponseDto>>

    @POST("api/auth/refresh")
    suspend fun refresh(@Body request: RefreshRequestDto): Response<ApiEnvelope<TokenPairDto>>

    @POST("api/auth/logout")
    suspend fun logout(@Body request: LogoutRequestDto): Response<ApiEnvelope<Map<String, Boolean>>>

    @GET("api/user/info")
    suspend fun getUserInfo(@Header("Authorization") authorization: String): Response<ApiEnvelope<UserDto>>

    @GET("api/imports")
    suspend fun listImports(@Header("Authorization") authorization: String): Response<ApiEnvelope<List<ImportSummaryDto>>>

    @POST("api/imports/upload")
    suspend fun uploadImport(
        @Header("Authorization") authorization: String,
        @Body requestBody: RequestBody
    ): Response<ApiEnvelope<ImportUploadResponseDto>>

    @Streaming
    @GET("api/imports/{importId}/export")
    suspend fun exportImport(
        @Header("Authorization") authorization: String,
        @Path("importId") importId: String
    ): Response<ResponseBody>
}
