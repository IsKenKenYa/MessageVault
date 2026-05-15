package com.iskenkenya.commory.sdk.storage.dao

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import com.iskenkenya.commory.sdk.storage.entity.ContactsEntity

@Dao
interface ContactDao {

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertContacts(contacts: List<ContactsEntity>)

    @Query("SELECT * FROM contacts ORDER BY name ASC")
    suspend fun getAllContacts(): List<ContactsEntity>

    @Query("DELETE FROM contacts")
    suspend fun deleteAllContacts()

    @Query("SELECT COUNT(id) FROM contacts")
    suspend fun getContactsCount(): Int
}
