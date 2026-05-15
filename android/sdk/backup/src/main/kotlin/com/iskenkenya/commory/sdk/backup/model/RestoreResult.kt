package com.iskenkenya.commory.sdk.backup.model

data class RestoreResult(
    val success: Boolean,
    val message: String,
    val restoredSmsCount: Int = 0,
    val restoredCallLogsCount: Int = 0,
    val restoredContactsCount: Int = 0
)
