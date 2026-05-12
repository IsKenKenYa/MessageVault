package imken.messagevault.sdk.backup.writer

import imken.messagevault.sdk.backup.model.SmsData

interface SmsWriter {
    suspend fun write(messages: List<SmsData>): Int
}
