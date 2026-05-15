package com.iskenkenya.commory.sdk.backup.reader

import com.iskenkenya.commory.sdk.backup.model.BackupReadData

interface BackupFileReader {
    suspend fun read(filePath: String): BackupReadData?
}
