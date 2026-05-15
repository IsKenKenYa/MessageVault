package com.iskenkenya.commory.sdk.backup

import com.iskenkenya.commory.sdk.backup.model.BackupData
import com.iskenkenya.commory.sdk.backup.model.BackupResult
import com.iskenkenya.commory.sdk.backup.model.CallLog
import com.iskenkenya.commory.sdk.backup.model.Contact
import com.iskenkenya.commory.sdk.backup.model.Message
import com.iskenkenya.commory.sdk.backup.model.BackupWriteStats
import com.iskenkenya.commory.sdk.backup.msglayer.MsgLayerMapper
import com.iskenkenya.commory.sdk.backup.reader.BackupFileReader
import com.iskenkenya.commory.sdk.backup.reader.CallLogReader
import com.iskenkenya.commory.sdk.backup.reader.ContactReader
import com.iskenkenya.commory.sdk.backup.reader.SmsReader
import com.iskenkenya.commory.sdk.backup.serializer.BackupSerializer
import com.iskenkenya.commory.sdk.backup.writer.BackupFileWriter
import com.iskenkenya.commory.sdk.backup.writer.CallLogWriter
import com.iskenkenya.commory.sdk.backup.writer.ContactWriter
import com.iskenkenya.commory.sdk.backup.writer.SmsWriter

class BackupManager(
    private val smsReader: SmsReader,
    private val callLogReader: CallLogReader,
    private val contactReader: ContactReader,
    private val backupFileWriter: BackupFileWriter,
    private val backupFileReader: BackupFileReader,
    private val smsWriter: SmsWriter,
    private val callLogWriter: CallLogWriter,
    private val contactWriter: ContactWriter,
    private val serializer: BackupSerializer = BackupSerializer(),
    private val msgLayerMapper: MsgLayerMapper = MsgLayerMapper()
) {

    suspend fun performBackup(
        hasSmsPermission: Boolean,
        hasCallLogPermission: Boolean,
        hasContactsPermission: Boolean,
        deviceInfo: String = "",
        deviceId: String = "",
        appVersion: String = ""
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

        val export = msgLayerMapper.toRootExport(
            messages = messages.orEmpty(),
            callLogs = callLogs.orEmpty(),
            contacts = contacts.orEmpty(),
            deviceInfo = deviceInfo,
            deviceId = deviceId.ifBlank { "unknown-device" },
            appVersion = appVersion.ifBlank { "unknown" }
        )

        return backupFileWriter.writeBackup(
            export = export,
            stats = BackupWriteStats(
                smsCount = messagesCount,
                callLogCount = callLogsCount,
                contactCount = contactsCount
            )
        )
    }

    suspend fun restoreFromFile(filePath: String): BackupData? {
        val backupReadData = backupFileReader.read(filePath) ?: return null
        return BackupData(
            messages = backupReadData.messages.map { message ->
                Message(
                    id = message.id,
                    address = message.address,
                    body = message.body,
                    date = message.date,
                    type = message.type,
                    readState = message.readState,
                    messageStatus = message.messageStatus,
                    threadId = message.threadId
                )
            },
            callLogs = backupReadData.callLogs.map { callLog ->
                CallLog(
                    id = callLog.id,
                    number = callLog.number,
                    type = callLog.type,
                    date = callLog.date,
                    duration = callLog.duration,
                    contact = callLog.contact
                )
            },
            contacts = backupReadData.contacts.map { contact ->
                Contact(
                    id = contact.id,
                    name = contact.name,
                    phoneNumbers = contact.phoneNumbers.toMutableList(),
                    emails = contact.emails
                )
            },
            timestamp = backupReadData.timestamp,
            deviceInfo = backupReadData.deviceInfo
        )
    }

    suspend fun restoreSms(messages: List<com.iskenkenya.commory.sdk.backup.model.Message>): Int {
        return smsWriter.write(messages.map { message ->
            com.iskenkenya.commory.sdk.backup.model.SmsData(
                id = message.id,
                address = message.address,
                body = message.body,
                date = message.date,
                type = message.type,
                readState = message.readState,
                messageStatus = message.messageStatus,
                threadId = message.threadId
            )
        })
    }

    suspend fun restoreCallLogs(callLogs: List<com.iskenkenya.commory.sdk.backup.model.CallLog>): Int {
        return callLogWriter.write(callLogs.map { callLog ->
            com.iskenkenya.commory.sdk.backup.model.CallLogData(
                id = callLog.id,
                number = callLog.number,
                type = callLog.type,
                date = callLog.date,
                duration = callLog.duration,
                contact = callLog.contact
            )
        })
    }

    suspend fun restoreContacts(contacts: List<com.iskenkenya.commory.sdk.backup.model.Contact>): Int {
        return contactWriter.write(contacts.map { contact ->
            com.iskenkenya.commory.sdk.backup.model.ContactData(
                id = contact.id,
                name = contact.name,
                phoneNumbers = contact.phoneNumbers,
                emails = contact.emails
            )
        })
    }

    fun getSerializer(): BackupSerializer = serializer
}
