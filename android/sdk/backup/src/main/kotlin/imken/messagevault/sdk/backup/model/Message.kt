package imken.messagevault.sdk.backup.model

import com.google.gson.annotations.Expose
import com.google.gson.annotations.SerializedName

data class Message(
    @Expose
    @SerializedName("id")
    val id: Long = 0,

    @Expose
    @SerializedName("addr")
    val address: String,

    @Expose
    @SerializedName("body")
    val body: String? = "",

    @Expose
    @SerializedName("date")
    val date: Long,

    @Expose
    @SerializedName("type")
    val type: Int,

    @SerializedName("read")
    val readState: Int? = 0,

    @SerializedName("status")
    val messageStatus: Int? = 0,

    @SerializedName("thread_id")
    val threadId: Long = 0
)
