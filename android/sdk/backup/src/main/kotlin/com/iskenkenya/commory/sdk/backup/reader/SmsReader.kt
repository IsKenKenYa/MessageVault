package com.iskenkenya.commory.sdk.backup.reader

import com.iskenkenya.commory.sdk.backup.model.Message

interface SmsReader {
    suspend fun readSms(): List<Message>?
}
