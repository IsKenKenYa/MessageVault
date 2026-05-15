package com.iskenkenya.commory.sdk.backup.writer

import com.iskenkenya.commory.sdk.backup.model.BackupResult
import com.iskenkenya.commory.sdk.backup.model.BackupWriteStats
import com.iskenkenya.commory.sdk.backup.msglayer.model.MsgLayerRootExport

interface BackupFileWriter {
    suspend fun writeBackup(export: MsgLayerRootExport, stats: BackupWriteStats): BackupResult
}
