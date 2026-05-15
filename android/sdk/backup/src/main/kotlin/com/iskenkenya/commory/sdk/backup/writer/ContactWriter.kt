package com.iskenkenya.commory.sdk.backup.writer

import com.iskenkenya.commory.sdk.backup.model.ContactData

interface ContactWriter {
    suspend fun write(contacts: List<ContactData>): Int
}
