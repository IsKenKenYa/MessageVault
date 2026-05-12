package imken.messagevault.sdk.backup.model

data class BackupResult(
    val success: Boolean,
    val timestamp: Long,
    val appVersion: String,
    val deviceId: String,
    val smsCount: Int = 0,
    val callLogCount: Int = 0,
    val fileName: String? = null,
    val filePath: String? = null,
    val errorMessage: String? = null
)
