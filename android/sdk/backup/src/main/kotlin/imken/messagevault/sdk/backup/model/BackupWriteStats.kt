package imken.messagevault.sdk.backup.model

data class BackupWriteStats(
    val smsCount: Int = 0,
    val callLogCount: Int = 0,
    val contactCount: Int = 0
)
