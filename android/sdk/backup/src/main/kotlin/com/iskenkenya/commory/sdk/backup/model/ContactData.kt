package com.iskenkenya.commory.sdk.backup.model

data class ContactData(
    val id: Long,
    val name: String,
    val phoneNumbers: List<String>,
    val emails: List<String>? = null
)
