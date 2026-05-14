package imken.messagevault.sdk.backup.writer

import imken.messagevault.sdk.backup.model.BackupResult
import imken.messagevault.sdk.backup.model.BackupWriteStats
import imken.messagevault.sdk.backup.msglayer.model.MsgLayerRootExport

interface BackupFileWriter {
    suspend fun writeBackup(export: MsgLayerRootExport, stats: BackupWriteStats): BackupResult
}
