package com.iskenkenya.commory.sdk.backup.model

import com.google.gson.annotations.Expose
import com.google.gson.annotations.SerializedName

data class CallLog(
    @Expose
    @SerializedName("id")
    val id: Long,

    @Expose
    @SerializedName("num")
    val number: String,

    @Expose
    @SerializedName("type")
    val type: Int,

    @Expose
    @SerializedName("date")
    val date: Long,

    @Expose
    @SerializedName("dur")
    val duration: Int,

    @Expose
    @SerializedName("name")
    val contact: String? = null
)
