package imken.messagevault.sdk.backup.writer

import imken.messagevault.sdk.backup.model.ContactData

interface ContactWriter {
    suspend fun write(contacts: List<ContactData>): Int
}
