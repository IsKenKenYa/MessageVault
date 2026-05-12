package imken.messagevault.sdk.backup.serializer

import com.google.gson.FieldNamingPolicy
import com.google.gson.GsonBuilder
import com.google.gson.reflect.TypeToken
import imken.messagevault.sdk.backup.model.BackupData
import imken.messagevault.sdk.backup.model.CallLog
import imken.messagevault.sdk.backup.model.Contact
import imken.messagevault.sdk.backup.model.Message

class BackupSerializer {

    private val gson = GsonBuilder()
        .setFieldNamingPolicy(FieldNamingPolicy.LOWER_CASE_WITH_UNDERSCORES)
        .serializeNulls()
        .disableHtmlEscaping()
        .setLenient()
        .setPrettyPrinting()
        .create()

    fun toJson(backupData: BackupData): String {
        return gson.toJson(backupData)
    }

    fun fromJson(json: String): BackupData? {
        return try {
            val typeToken = object : TypeToken<BackupData>() {}.type
            gson.fromJson<BackupData>(json, typeToken)
        } catch (e: Exception) {
            null
        }
    }

    fun getGson() = gson

    fun cleanMessage(message: Message): Message {
        return Message(
            id = message.id,
            address = message.address.replace(Regex("[^\\p{Print}]"), ""),
            body = message.body?.replace(Regex("[^\\p{Print}]"), ""),
            date = message.date,
            type = message.type,
            readState = message.readState,
            messageStatus = message.messageStatus,
            threadId = message.threadId
        )
    }

    fun cleanCallLog(callLog: CallLog): CallLog {
        return CallLog(
            id = callLog.id,
            number = callLog.number.replace(Regex("[^\\p{Print}]"), ""),
            type = callLog.type,
            date = callLog.date,
            duration = callLog.duration,
            contact = callLog.contact?.replace(Regex("[^\\p{Print}]"), "")
        )
    }

    fun cleanContact(contact: Contact): Contact {
        return Contact(
            id = contact.id,
            name = contact.name.replace(Regex("[^\\p{Print}]"), ""),
            phoneNumbers = contact.phoneNumbers.map { it.replace(Regex("[^\\p{Print}]"), "") }.toMutableList(),
            emails = contact.emails?.map { it.replace(Regex("[^\\p{Print}]"), "") },
            addresses = contact.addresses?.map {
                Contact.Address(
                    type = it.type.replace(Regex("[^\\p{Print}]"), ""),
                    value = it.value.replace(Regex("[^\\p{Print}]"), "")
                )
            },
            groups = contact.groups?.map { it.replace(Regex("[^\\p{Print}]"), "") },
            note = contact.note?.replace(Regex("[^\\p{Print}]"), ""),
            websites = contact.websites?.map { it.replace(Regex("[^\\p{Print}]"), "") },
            events = contact.events?.map {
                Contact.Event(
                    type = it.type.replace(Regex("[^\\p{Print}]"), ""),
                    date = it.date.replace(Regex("[^\\p{Print}]"), "")
                )
            },
            relationships = contact.relationships?.map {
                Contact.Relationship(
                    type = it.type.replace(Regex("[^\\p{Print}]"), ""),
                    name = it.name.replace(Regex("[^\\p{Print}]"), "")
                )
            },
            socialProfiles = contact.socialProfiles?.map {
                Contact.SocialProfile(
                    type = it.type.replace(Regex("[^\\p{Print}]"), ""),
                    value = it.value.replace(Regex("[^\\p{Print}]"), "")
                )
            }
        )
    }

    fun cleanBackupData(backupData: BackupData): BackupData {
        val cleanedMessages = backupData.messages?.map { cleanMessage(it) }
        val cleanedCallLogs = backupData.callLogs?.map { cleanCallLog(it) }
        val cleanedContacts = backupData.contacts?.map { cleanContact(it) }

        return BackupData(
            messages = cleanedMessages,
            callLogs = cleanedCallLogs,
            contacts = cleanedContacts,
            timestamp = backupData.timestamp,
            deviceInfo = backupData.deviceInfo
        )
    }

    fun serializeWithSizeLimit(backupData: BackupData, maxBytes: Int = 10 * 1024 * 1024): String? {
        val cleanedData = cleanBackupData(backupData)

        val testJson = toJson(BackupData(
            messages = emptyList(),
            callLogs = emptyList(),
            contacts = emptyList(),
            timestamp = cleanedData.timestamp,
            deviceInfo = cleanedData.deviceInfo
        ))
        if (testJson.isBlank() || testJson == "{}" || testJson == "null") {
            return null
        }

        var finalMessages = cleanedData.messages
        var finalCallLogs = cleanedData.callLogs
        var finalContacts = cleanedData.contacts

        if (!cleanedData.messages.isNullOrEmpty()) {
            try {
                gson.toJson(cleanedData.messages)
            } catch (e: Exception) {
                finalMessages = emptyList()
            }
        }

        if (!cleanedData.callLogs.isNullOrEmpty()) {
            try {
                gson.toJson(cleanedData.callLogs)
            } catch (e: Exception) {
                finalCallLogs = emptyList()
            }
        }

        if (!cleanedData.contacts.isNullOrEmpty()) {
            try {
                gson.toJson(cleanedData.contacts)
            } catch (e: Exception) {
                finalContacts = emptyList()
            }
        }

        val finalData = BackupData(
            messages = finalMessages,
            callLogs = finalCallLogs,
            contacts = finalContacts,
            timestamp = cleanedData.timestamp,
            deviceInfo = cleanedData.deviceInfo
        )

        val jsonString: String
        try {
            jsonString = toJson(finalData)
            if (jsonString.isBlank() || jsonString == "{}" || jsonString == "null") {
                return null
            }
            fromJson(jsonString) ?: return null
        } catch (e: Exception) {
            return null
        }

        val jsonBytes = jsonString.toByteArray(Charsets.UTF_8)
        if (jsonBytes.size > maxBytes) {
            val limitedData = BackupData(
                messages = finalMessages?.take(500),
                callLogs = finalCallLogs?.take(500),
                contacts = finalContacts?.take(500),
                timestamp = cleanedData.timestamp,
                deviceInfo = "${cleanedData.deviceInfo} (数据已精简，原始数据: ${finalMessages?.size ?: 0}条短信, ${finalCallLogs?.size ?: 0}条通话记录, ${finalContacts?.size ?: 0}个联系人)"
            )
            return toJson(limitedData)
        }

        return jsonString
    }
}
