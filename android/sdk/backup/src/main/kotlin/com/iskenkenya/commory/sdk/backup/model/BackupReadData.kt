package com.iskenkenya.commory.sdk.backup.model

data class BackupReadData(
    val messages: List<SmsData> = emptyList(),
    val callLogs: List<CallLogData> = emptyList(),
    val contacts: List<ContactData> = emptyList(),
    val timestamp: Long = System.currentTimeMillis(),
    val deviceInfo: String = ""
)
