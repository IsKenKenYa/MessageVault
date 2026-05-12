package imken.messagevault.sdk.backup.reader

import imken.messagevault.sdk.backup.model.Contact

interface ContactReader {
    suspend fun readContacts(): List<Contact>?
}
