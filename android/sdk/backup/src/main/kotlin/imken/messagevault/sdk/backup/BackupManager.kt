package imken.messagevault.sdk.backup

import imken.messagevault.sdk.backup.model.BackupData
import imken.messagevault.sdk.backup.model.BackupResult
import imken.messagevault.sdk.backup.reader.BackupFileReader
import imken.messagevault.sdk.backup.reader.CallLogReader
import imken.messagevault.sdk.backup.reader.ContactReader
import imken.messagevault.sdk.backup.reader.SmsReader
import imken.messagevault.sdk.backup.serializer.BackupSerializer
import imken.messagevault.sdk.backup.writer.BackupFileWriter
import imken.messagevault.sdk.backup.writer.CallLogWriter
import imken.messagevault.sdk.backup.writer.ContactWriter
import imken.messagevault.sdk.backup.writer.SmsWriter

class BackupManager(
    private val smsReader: SmsReader,
    private val callLogReader: CallLogReader,
    private val contactReader: ContactReader,
    private val backupFileWriter: BackupFileWriter,
    private val backupFileReader: BackupFileReader,
    private val smsWriter: SmsWriter,
    private val callLogWriter: CallLogWriter,
    private val contactWriter: ContactWriter,
    private val serializer: BackupSerializer = BackupSerializer()
) {

    suspend fun performBackup(
        hasSmsPermission: Boolean,
        hasCallLogPermission: Boolean,
        hasContactsPermission: Boolean,
        deviceInfo: String = ""
    ): BackupResult {
        if (!hasSmsPermission && !hasCallLogPermission && !hasContactsPermission) {
            return BackupResult(
                success = false,
                timestamp = System.currentTimeMillis(),
                appVersion = "",
                deviceId = "",
                errorMessage = "备份失败: 没有任何所需权限"
            )
        }

        val messages = if (hasSmsPermission) smsReader.readSms() else null
        val messagesCount = messages?.size ?: 0

        val callLogs = if (hasCallLogPermission) callLogReader.readCallLogs() else null
        val callLogsCount = callLogs?.size ?: 0

        val contacts = if (hasContactsPermission) contactReader.readContacts() else null
        val contactsCount = contacts?.size ?: 0

        if (messagesCount == 0 && callLogsCount == 0 && contactsCount == 0) {
            return BackupResult(
                success = false,
                timestamp = System.currentTimeMillis(),
                appVersion = "",
                deviceId = "",
                smsCount = 0,
                callLogCount = 0,
                errorMessage = "没有找到任何数据可以备份"
            )
        }

        val backupData = BackupData(
            messages = messages,
            callLogs = callLogs,
            contacts = contacts,
            timestamp = System.currentTimeMillis(),
            deviceInfo = deviceInfo
        )

        val json = serializer.serializeWithSizeLimit(backupData)
            ?: return BackupResult(
                success = false,
                timestamp = System.currentTimeMillis(),
                appVersion = "",
                deviceId = "",
                smsCount = messagesCount,
                callLogCount = callLogsCount,
                errorMessage = "创建备份文件失败"
            )

        return backupFileWriter.writeBackup(backupData, "")
    }

    suspend fun restoreFromFile(filePath: String): BackupData? {
        return backupFileReader.readBackup(filePath)
    }

    suspend fun restoreSms(messages: List<imken.messagevault.sdk.backup.model.Message>): Int {
        return smsWriter.writeSms(messages)
    }

    suspend fun restoreCallLogs(callLogs: List<imken.messagevault.sdk.backup.model.CallLog>): Int {
        return callLogWriter.writeCallLogs(callLogs)
    }

    suspend fun restoreContacts(contacts: List<imken.messagevault.sdk.backup.model.Contact>): Int {
        return contactWriter.writeContacts(contacts)
    }

    fun getSerializer(): BackupSerializer = serializer
}
