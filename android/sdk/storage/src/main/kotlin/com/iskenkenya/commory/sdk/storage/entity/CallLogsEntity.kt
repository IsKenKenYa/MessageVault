package com.iskenkenya.commory.sdk.storage.entity

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "call_logs")
data class CallLogsEntity(
    @PrimaryKey(autoGenerate = true)
    val id: Long = 0,
    val number: String,
    val name: String,
    val date: Long,
    val duration: Int,
    val type: Int
)
