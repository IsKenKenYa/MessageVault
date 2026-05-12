package imken.messagevault.sdk.storage

interface RemoteStorageProvider {
    val isConnected: Boolean

    suspend fun upload(data: ByteArray, remotePath: String): StorageResult
    suspend fun download(remotePath: String): StorageResult
    suspend fun listRemote(prefix: String = ""): List<StorageFileInfo>
    suspend fun deleteRemote(remotePath: String): StorageResult
    suspend fun connect(config: RemoteStorageConfig): Boolean
    suspend fun disconnect()
}

data class RemoteStorageConfig(
    val serverUrl: String,
    val authToken: String? = null,
    val maxRetryCount: Int = 3,
    val timeoutMs: Long = 30000
)
