package imken.messagevault.mobile.data.backup

import android.content.Context
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import imken.messagevault.mobile.model.BackupData
import imken.messagevault.sdk.backup.model.BackupReadData
import imken.messagevault.sdk.backup.model.CallLogData
import imken.messagevault.sdk.backup.model.ContactData
import imken.messagevault.sdk.backup.model.SmsData
import imken.messagevault.sdk.backup.reader.BackupFileReader
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import timber.log.Timber
import java.io.File
import java.io.FileReader

class AndroidBackupFileReader(
    private val context: Context
) : BackupFileReader {

    private val gson = Gson()

    override suspend fun read(filePath: String): BackupReadData? = withContext(Dispatchers.IO) {
        try {
            val file = File(filePath)
            if (!file.exists() || !file.isFile || !file.canRead()) {
                Timber.e("[Mobile] ERROR [Restore] Backup file does not exist or is not readable: $filePath")
                return@withContext null
            }

            FileReader(file).use { reader ->
                val fileContent = file.readText()
                val containsCallLogs = fileContent.contains("\"callLogs\"") || fileContent.contains("\"call_logs\"")
                val containsMessages = fileContent.contains("\"messages\"") || fileContent.contains("\"sms\"")
                val containsContacts = fileContent.contains("\"contacts\"")

                Timber.d("[Mobile] DEBUG [Restore] Backup file content analysis: hasCallLogs=$containsCallLogs, hasMessages=$containsMessages, hasContacts=$containsContacts")

                val typeToken = object : TypeToken<BackupData>() {}.type
                val backupData = gson.fromJson<BackupData>(reader, typeToken)
                if (backupData == null) {
                    Timber.e("[Mobile] ERROR [Restore] Parsed backup file returned null: $filePath")
                    return@withContext null
                }

                val messagesCount = backupData.messages?.size ?: 0
                val callLogsCount = backupData.callLogs?.size ?: 0
                val contactsCount = backupData.contacts?.size ?: 0

                Timber.i("[Mobile] INFO [Restore] Successfully parsed backup file: SMS=$messagesCount, callLogs=$callLogsCount, contacts=$contactsCount")

                if (containsCallLogs && callLogsCount == 0) {
                    Timber.w("[Mobile] WARN [Restore] File contains call log fields but parsed as empty, possible field name mismatch")
                }

                val smsDataList = backupData.messages?.map { msg ->
                    SmsData(
                        id = msg.id,
                        address = msg.address,
                        body = msg.body,
                        date = msg.date,
                        type = msg.type,
                        readState = msg.readState,
                        messageStatus = msg.messageStatus,
                        threadId = msg.threadId
                    )
                } ?: emptyList()

                val callLogDataList = backupData.callLogs?.map { callLog ->
                    CallLogData(
                        id = callLog.id,
                        number = callLog.number,
                        type = callLog.type,
                        date = callLog.date,
                        duration = callLog.duration,
                        contact = callLog.contact
                    )
                } ?: emptyList()

                val contactDataList = backupData.contacts?.map { contact ->
                    ContactData(
                        id = contact.id,
                        name = contact.name,
                        phoneNumbers = contact.phoneNumbers.toList(),
                        emails = contact.emails
                    )
                } ?: emptyList()

                BackupReadData(
                    messages = smsDataList,
                    callLogs = callLogDataList,
                    contacts = contactDataList,
                    timestamp = backupData.timestamp,
                    deviceInfo = backupData.deviceInfo
                )
            }
        } catch (e: Exception) {
            Timber.e(e, "[Mobile] ERROR [Restore] Failed to parse backup file: $filePath, ${e.message}")
            null
        }
    }
}
