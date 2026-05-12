package imken.messagevault.sdk.backup.reader

import imken.messagevault.sdk.backup.model.Message

interface SmsReader {
    suspend fun readSms(): List<Message>?
}
