package com.iskenkenya.commory.sdk.storage.entity

import android.os.Parcel
import android.os.Parcelable
import androidx.room.Entity
import androidx.room.PrimaryKey
import com.google.gson.annotations.SerializedName

@Entity(tableName = "contacts")
data class ContactsEntity(
    @PrimaryKey
    @SerializedName("id")
    val id: Long,
    @SerializedName("name")
    val name: String,
    @SerializedName("phone_numbers")
    val phoneNumbers: List<String>,
    @SerializedName("emails")
    val emails: List<String>,
    @SerializedName("photo_uri")
    val photoUri: String? = null,
    @SerializedName("last_updated")
    val lastUpdated: Long = System.currentTimeMillis()
) : Parcelable {
    constructor(parcel: Parcel) : this(
        parcel.readLong(),
        parcel.readString() ?: "",
        parcel.createStringArrayList() ?: emptyList(),
        parcel.createStringArrayList() ?: emptyList()
    )

    override fun writeToParcel(parcel: Parcel, flags: Int) {
        parcel.writeLong(id)
        parcel.writeString(name)
        parcel.writeStringList(phoneNumbers)
        parcel.writeStringList(emails)
    }

    override fun describeContents(): Int {
        return 0
    }

    companion object CREATOR : Parcelable.Creator<ContactsEntity> {
        override fun createFromParcel(parcel: Parcel): ContactsEntity {
            return ContactsEntity(parcel)
        }

        override fun newArray(size: Int): Array<ContactsEntity?> {
            return arrayOfNulls(size)
        }
    }
}
