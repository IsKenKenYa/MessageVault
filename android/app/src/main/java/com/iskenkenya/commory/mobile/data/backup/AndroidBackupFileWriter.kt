package com.iskenkenya.commory.mobile.data.backup

import android.content.Context
import android.os.Build
import com.iskenkenya.commory.sdk.backup.model.BackupResult
import com.iskenkenya.commory.sdk.backup.model.BackupWriteStats
import com.iskenkenya.commory.sdk.backup.msglayer.MsgLayerSerializer
import com.iskenkenya.commory.sdk.backup.msglayer.model.MSG_LAYER_VERSION
import com.iskenkenya.commory.sdk.backup.msglayer.model.MsgLayerRootExport
import com.iskenkenya.commory.sdk.backup.writer.BackupFileWriter
import timber.log.Timber
import java.io.File
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

class AndroidBackupFileWriter(private val context: Context) : BackupFileWriter {

    private val serializer = MsgLayerSerializer()

    override suspend fun writeBackup(export: MsgLayerRootExport, stats: BackupWriteStats): BackupResult {
        try {
            val backupDir = File(context.getExternalFilesDir(null), "backups")
            if (!backupDir.exists()) {
                backupDir.mkdirs()
            }

            val fileName = generateUserFriendlyFileName()
            val backupFile = File(backupDir, fileName)

            Timber.d(
                "[Mobile] DEBUG [Backup] MsgLayer数据准备情况: 短信=${stats.smsCount}, 通话记录=${stats.callLogCount}, 联系人=${stats.contactCount}, 版本=${export.version}"
            )

            val jsonString = serializer.toJson(export)

            try {
                backupFile.writeText(jsonString)
                Timber.d("[Mobile] DEBUG [Backup] 成功将数据写入文件")
            } catch (e: Exception) {
                Timber.e(e, "[Mobile] ERROR [Backup] 写入文件失败: ${e.message}")
                return BackupResult(
                    success = false,
                    timestamp = System.currentTimeMillis(),
                    appVersion = "",
                    deviceId = "",
                    errorMessage = "写入文件失败: ${e.message}"
                )
            }

            if (!backupFile.exists() || backupFile.length() <= 10) {
                Timber.e("[Mobile] ERROR [Backup] 创建的备份文件为空或过小: ${backupFile.length()} 字节")
                backupFile.delete()
                return BackupResult(
                    success = false,
                    timestamp = System.currentTimeMillis(),
                    appVersion = "",
                    deviceId = "",
                    errorMessage = "创建的备份文件为空或过小"
                )
            }

            Timber.i("[Mobile] INFO [Backup] 成功创建备份文件: ${backupFile.absolutePath}, 大小: ${backupFile.length()} 字节")

            return BackupResult(
                success = true,
                timestamp = System.currentTimeMillis(),
                appVersion = export.source.appVersion,
                deviceId = export.source.deviceId,
                smsCount = stats.smsCount,
                callLogCount = stats.callLogCount,
                fileName = backupFile.name,
                filePath = backupFile.absolutePath
            )

        } catch (e: Exception) {
            Timber.e(e, "[Mobile] ERROR [Backup] 创建本地备份文件失败; Context: ${e.message}")
            return BackupResult(
                success = false,
                timestamp = System.currentTimeMillis(),
                appVersion = "",
                deviceId = "",
                errorMessage = "创建本地备份文件失败: ${e.message}"
            )
        }
    }

    private fun generateUserFriendlyFileName(deviceName: String? = null): String {
        val dateFormat = SimpleDateFormat("yyyy-MM-dd_HH-mm", Locale.getDefault())
        val timestamp = dateFormat.format(Date())
        val device = deviceName ?: Build.MODEL.replace(" ", "_")
        return "${MSG_LAYER_VERSION.replace("/", "-")}_${device}_${timestamp}.json"
    }
}
