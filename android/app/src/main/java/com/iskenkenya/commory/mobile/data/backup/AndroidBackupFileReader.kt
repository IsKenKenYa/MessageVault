package com.iskenkenya.commory.mobile.data.backup

import android.content.Context
import com.google.gson.Gson
import com.google.gson.JsonObject
import com.google.gson.JsonParser
import com.google.gson.reflect.TypeToken
import com.iskenkenya.commory.mobile.model.BackupData
import com.iskenkenya.commory.sdk.backup.msglayer.MsgLayerSerializer
import com.iskenkenya.commory.sdk.backup.msglayer.model.MsgLayerIdentity
import com.iskenkenya.commory.sdk.backup.msglayer.model.MsgLayerRootExport
import com.iskenkenya.commory.sdk.backup.model.BackupReadData
import com.iskenkenya.commory.sdk.backup.model.CallLogData
import com.iskenkenya.commory.sdk.backup.model.ContactData
import com.iskenkenya.commory.sdk.backup.model.SmsData
import com.iskenkenya.commory.sdk.backup.reader.BackupFileReader
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import timber.log.Timber
import java.io.File
import java.io.FileReader
import kotlin.math.absoluteValue

class AndroidBackupFileReader(
    private val context: Context
) : BackupFileReader {

    private val gson = Gson()
    private val msgLayerGson = MsgLayerSerializer().gson()

    override suspend fun read(filePath: String): BackupReadData? = withContext(Dispatchers.IO) {
        try {
            val file = File(filePath)
            if (!file.exists() || !file.isFile || !file.canRead()) {
                Timber.e("[Mobile] ERROR [Restore] Backup file does not exist or is not readable: $filePath")
                return@withContext null
            }

            val fileContent = file.readText()
            val containsCallLogs = fileContent.contains("\"callLogs\"") || fileContent.contains("\"call_logs\"")
            val containsMessages = fileContent.contains("\"messages\"") || fileContent.contains("\"sms\"")
            val containsContacts = fileContent.contains("\"contacts\"")

            Timber.d("[Mobile] DEBUG [Restore] Backup file content analysis: hasCallLogs=$containsCallLogs, hasMessages=$containsMessages, hasContacts=$containsContacts")

            val parsedRoot = runCatching {
                JsonParser.parseString(fileContent).asJsonObject
            }.getOrNull()

            if (parsedRoot?.get("version")?.asString == com.iskenkenya.commory.sdk.backup.msglayer.model.MSG_LAYER_VERSION) {
                val msgLayer = msgLayerGson.fromJson(fileContent, MsgLayerRootExport::class.java)
                    ?: return@withContext null
                return@withContext msgLayer.toBackupReadData()
            }

            FileReader(file).use { reader ->
                val typeToken = object : TypeToken<BackupData>() {}.type
                val backupData = gson.fromJson<BackupData>(reader, typeToken)
                    ?: return@withContext null

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

    private fun MsgLayerRootExport.toBackupReadData(): BackupReadData {
        val identitiesById = identities.associateBy { it.id }
        val messages = events.filter { it.type == "sms" }.mapNotNull { event ->
            val address = event.counterpartyAddress(identitiesById)
            val threadId = event.relations.firstOrNull { it.type == "same_thread" }
                ?.target
                ?.removePrefix("thread_")
                ?.toLongOrNull()
                ?: 0L
            SmsData(
                id = event.id.removePrefix("sms_").toLongOrNull() ?: event.id.hashCode().toLong().absoluteValue,
                address = address,
                body = event.content["text"] as? String ?: "",
                date = event.timestamp.toEpochMillis(),
                type = if (event.direction == "outbound") 2 else 1,
                readState = if ((event.meta["read"] as? Boolean) == true) 1 else 0,
                messageStatus = (event.meta["status"] as? Number)?.toInt() ?: 0,
                threadId = threadId
            )
        }
        val calls = events.filter { it.type == "call" }.mapNotNull { event ->
            val address = event.counterpartyAddress(identitiesById)
            CallLogData(
                id = event.id.removePrefix("call_").toLongOrNull() ?: event.id.hashCode().toLong().absoluteValue,
                number = address,
                type = when (event.content["call_type"] as? String ?: "unknown") {
                    "outgoing" -> 2
                    "incoming" -> 3
                    "rejected" -> 5
                    "voicemail" -> 4
                    else -> 1
                },
                date = event.timestamp.toEpochMillis(),
                duration = (event.content["duration_sec"] as? Number)?.toInt() ?: 0,
                contact = event.counterpartyName(identitiesById)
            )
        }
        val contacts = identities
            .filter { it.type == "person" }
            .map { identity ->
                ContactData(
                    id = identity.id.hashCode().toLong().absoluteValue,
                    name = identity.displayName,
                    phoneNumbers = identity.phones,
                    emails = identity.emails
                )
            }
        return BackupReadData(
            messages = messages,
            callLogs = calls,
            contacts = contacts,
            timestamp = exportedAt.toEpochMillis(),
            deviceInfo = source.deviceId
        )
    }

    private fun com.iskenkenya.commory.sdk.backup.msglayer.model.MsgLayerEvent.counterpartyAddress(
        identitiesById: Map<String, MsgLayerIdentity>
    ): String {
        val identityId = participants.firstOrNull { !it.startsWith("self/") } ?: return ""
        val identity = identitiesById[identityId]
        return identity?.phones?.firstOrNull() ?: identity?.displayName ?: identityId
    }

    private fun com.iskenkenya.commory.sdk.backup.msglayer.model.MsgLayerEvent.counterpartyName(
        identitiesById: Map<String, MsgLayerIdentity>
    ): String? {
        val identityId = participants.firstOrNull { !it.startsWith("self/") } ?: return null
        return identitiesById[identityId]?.displayName
    }

    private fun String.toEpochMillis(): Long = try {
        java.time.OffsetDateTime.parse(this).toInstant().toEpochMilli()
    } catch (exception: Exception) {
        Timber.w(exception, "[Mobile] WARN [Restore] Invalid RFC3339 timestamp: %s", this)
        0L
    }
}
