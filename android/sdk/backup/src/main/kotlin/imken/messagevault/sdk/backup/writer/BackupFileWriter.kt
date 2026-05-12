package imken.messagevault.sdk.backup.writer

import imken.messagevault.sdk.backup.model.BackupData
import imken.messagevault.sdk.backup.model.BackupResult

interface BackupFileWriter {
    suspend fun writeBackup(data: BackupData, filePath: String): BackupResult
}
