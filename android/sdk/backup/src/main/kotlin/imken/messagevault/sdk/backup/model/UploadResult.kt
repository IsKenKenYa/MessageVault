package imken.messagevault.sdk.backup.model

data class UploadResult(
    val success: Boolean,
    val fileId: String? = null,
    val errorMessage: String? = null
)
