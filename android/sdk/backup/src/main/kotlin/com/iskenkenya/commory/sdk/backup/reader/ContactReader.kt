package com.iskenkenya.commory.sdk.backup.reader

import com.iskenkenya.commory.sdk.backup.model.Contact

interface ContactReader {
    suspend fun readContacts(): List<Contact>?
}
