package com.iskenkenya.commory.sdk.backup.model

data class RestoreOptions(
    val sms: Boolean = true,
    val callLogs: Boolean = true,
    val contacts: Boolean = true
)
