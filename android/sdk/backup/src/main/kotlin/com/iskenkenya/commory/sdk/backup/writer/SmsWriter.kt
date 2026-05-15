package com.iskenkenya.commory.sdk.backup.writer

import com.iskenkenya.commory.sdk.backup.model.SmsData

interface SmsWriter {
    suspend fun write(messages: List<SmsData>): Int
}
