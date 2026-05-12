package imken.messagevault.sdk.backup.model

data class SmsData(
    val id: Long = 0,
    val address: String,
    val body: String? = "",
    val date: Long,
    val type: Int,
    val readState: Int? = 0,
    val messageStatus: Int? = 0,
    val threadId: Long = 0
)
