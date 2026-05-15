package com.iskenkenya.commory.sdk.storage.database

import android.content.Context
import androidx.room.Database
import androidx.room.Room
import androidx.room.RoomDatabase
import com.iskenkenya.commory.sdk.storage.dao.CallLogDao
import com.iskenkenya.commory.sdk.storage.dao.ContactDao
import com.iskenkenya.commory.sdk.storage.dao.MessageDao
import com.iskenkenya.commory.sdk.storage.entity.CallLogsEntity
import com.iskenkenya.commory.sdk.storage.entity.ContactsEntity
import com.iskenkenya.commory.sdk.storage.entity.MessageEntity

@Database(
    entities = [
        CallLogsEntity::class,
        MessageEntity::class,
        ContactsEntity::class
    ],
    version = 1,
    exportSchema = false
)
abstract class AppDatabase : RoomDatabase() {

    abstract fun callLogDao(): CallLogDao
    abstract fun messageDao(): MessageDao
    abstract fun contactDao(): ContactDao

    companion object {
        @Volatile
        private var INSTANCE: AppDatabase? = null

        fun getInstance(context: Context): AppDatabase {
            return INSTANCE ?: synchronized(this) {
                val instance = Room.databaseBuilder(
                    context.applicationContext,
                    AppDatabase::class.java,
                    "message_vault.db"
                )
                .fallbackToDestructiveMigration()
                .build()
                INSTANCE = instance
                instance
            }
        }
    }
}
