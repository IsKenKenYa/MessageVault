package imken.messagevault.sdk.backup.reader

import imken.messagevault.sdk.backup.model.BackupReadData

interface BackupFileReader {
    suspend fun read(filePath: String): BackupReadData?
}
