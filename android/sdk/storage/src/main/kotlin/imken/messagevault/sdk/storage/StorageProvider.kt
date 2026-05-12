package imken.messagevault.sdk.storage

import kotlinx.coroutines.flow.StateFlow

interface StorageProvider {
    val isAvailable: StateFlow<Boolean>

    suspend fun save(data: ByteArray, path: String): StorageResult
    suspend fun load(path: String): StorageResult
    suspend fun list(prefix: String = ""): List<StorageFileInfo>
    suspend fun delete(path: String): StorageResult
    suspend fun getStorageInfo(): StorageInfo
}

data class StorageFileInfo(
    val path: String,
    val size: Long,
    val lastModified: Long,
    val mimeType: String? = null
)

data class StorageInfo(
    val totalSpace: Long,
    val usedSpace: Long,
    val availableSpace: Long
)

sealed class StorageResult {
    data class Success(val path: String, val data: ByteArray? = null) : StorageResult()
    data class Error(val code: String, val message: String) : StorageResult()
}
