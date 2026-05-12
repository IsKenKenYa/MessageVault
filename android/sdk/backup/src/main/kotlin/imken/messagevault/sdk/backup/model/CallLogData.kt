package imken.messagevault.sdk.backup.model

data class CallLogData(
    val id: Long,
    val number: String,
    val type: Int,
    val date: Long,
    val duration: Int,
    val contact: String? = null
)
