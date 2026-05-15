package com.iskenkenya.commory.sdk.backup.model

import com.google.gson.annotations.Expose
import com.google.gson.annotations.SerializedName

data class Contact(
    @Expose
    @SerializedName("id")
    val id: Long,

    @Expose
    @SerializedName("name")
    val name: String,

    @Expose
    @SerializedName("phones")
    val phoneNumbers: MutableList<String>,

    @Expose
    @SerializedName("emails")
    val emails: List<String>? = null,

    @Expose
    @SerializedName("addresses")
    val addresses: List<Address>? = null,

    val photoData: String? = null,

    @Expose
    @SerializedName("note")
    val note: String? = null,

    @Expose
    @SerializedName("groups")
    val groups: List<String>? = null,

    @Expose
    @SerializedName("websites")
    val websites: List<String>? = null,

    @Expose
    @SerializedName("events")
    val events: List<Event>? = null,

    @Expose
    @SerializedName("relationships")
    val relationships: List<Relationship>? = null,

    @Expose
    @SerializedName("social_profiles")
    val socialProfiles: List<SocialProfile>? = null
) {
    data class Address(
        @Expose
        @SerializedName("type")
        val type: String,

        @Expose
        @SerializedName("value")
        val value: String
    )

    data class Event(
        @Expose
        @SerializedName("type")
        val type: String,

        @Expose
        @SerializedName("date")
        val date: String
    )

    data class Relationship(
        @Expose
        @SerializedName("type")
        val type: String,

        @Expose
        @SerializedName("name")
        val name: String
    )

    data class SocialProfile(
        @Expose
        @SerializedName("type")
        val type: String,

        @Expose
        @SerializedName("value")
        val value: String
    )
}
