package com.iskenkenya.commory.sdk.backup.msglayer

import com.iskenkenya.commory.sdk.backup.model.CallLog
import com.iskenkenya.commory.sdk.backup.model.Contact
import com.iskenkenya.commory.sdk.backup.model.Message
import com.iskenkenya.commory.sdk.backup.msglayer.model.MsgLayerEvent
import com.iskenkenya.commory.sdk.backup.msglayer.model.MsgLayerIdentity
import com.iskenkenya.commory.sdk.backup.msglayer.model.MsgLayerRelation
import com.iskenkenya.commory.sdk.backup.msglayer.model.MsgLayerRootExport
import com.iskenkenya.commory.sdk.backup.msglayer.model.MsgLayerSource
import java.security.MessageDigest
import java.time.Instant
import java.time.ZoneOffset
import java.time.format.DateTimeFormatter
import kotlin.math.absoluteValue

class MsgLayerMapper {

    fun toRootExport(
        messages: List<Message>,
        callLogs: List<CallLog>,
        contacts: List<Contact>,
        deviceInfo: String,
        deviceId: String,
        appVersion: String,
        exportedAtMillis: Long = System.currentTimeMillis()
    ): MsgLayerRootExport {
        val exportedAt = exportedAtMillis.toRfc3339()
        val selfIdentity = MsgLayerIdentity(
            id = selfIdentityId(deviceId),
            type = "device",
            displayName = deviceInfo.ifBlank { deviceId },
            phones = emptyList(),
            emails = emptyList(),
            avatar = null,
            labels = listOf("self"),
            meta = mapOf("source" to "system")
        )

        val identitiesById = linkedMapOf(selfIdentity.id to selfIdentity)
        contacts.forEach { contact ->
            val identity = contact.toIdentity()
            identitiesById[identity.id] = identity
        }

        val events = mutableListOf<MsgLayerEvent>()
        identitiesById.values
            .filter { it.id != selfIdentity.id && it.type == "person" }
            .forEach { identity ->
                events += MsgLayerEvent(
                    id = "contact_snapshot_${identity.id}",
                    type = "contact_snapshot",
                    timestamp = exportedAt,
                    direction = "system",
                    participants = listOf(identity.id),
                    content = mapOf(
                        "identity_id" to identity.id,
                        "snapshot" to identity
                    ),
                    meta = mapOf("source" to "contacts"),
                    relations = listOf(MsgLayerRelation("references_identity", identity.id))
                )
            }

        messages.forEach { message ->
            val participantId = resolveParticipantIdentity(message.address, identitiesById)
            events += MsgLayerEvent(
                id = "sms_${message.id}",
                type = "sms",
                timestamp = message.date.toRfc3339(),
                direction = message.toSmsDirection(),
                participants = listOf(selfIdentity.id, participantId).distinct(),
                content = mapOf(
                    "text" to (message.body ?: ""),
                    "attachments" to emptyList<String>()
                ),
                meta = mapOf(
                    "read" to (message.readState == 1),
                    "status" to (message.messageStatus ?: 0),
                    "raw_type" to message.type
                ),
                relations = listOf(
                    MsgLayerRelation("same_thread", "thread_${message.threadId}"),
                    MsgLayerRelation("references_identity", participantId)
                )
            )
        }

        callLogs.forEach { callLog ->
            val participantId = resolveParticipantIdentity(callLog.number, identitiesById, callLog.contact)
            events += MsgLayerEvent(
                id = "call_${callLog.id}",
                type = "call",
                timestamp = callLog.date.toRfc3339(),
                direction = callLog.toCallDirection(),
                participants = listOf(selfIdentity.id, participantId).distinct(),
                content = mapOf(
                    "duration_sec" to callLog.duration,
                    "call_type" to callLog.toCallType(),
                    "recording" to null
                ),
                meta = mapOf(
                    "contact_name" to callLog.contact,
                    "raw_type" to callLog.type
                ),
                relations = listOf(MsgLayerRelation("references_identity", participantId))
            )
        }

        // String sorting is safe here because all timestamps are normalized to UTC RFC3339.
        return MsgLayerRootExport(
            exportedAt = exportedAt,
            source = MsgLayerSource(
                platform = "android",
                deviceId = deviceId,
                appVersion = appVersion
            ),
            identities = identitiesById.values.toList(),
            events = events.sortedByDescending { it.timestamp },
            indexes = mapOf(
                "timeline" to events.map { it.id }
            )
        )
    }

    private fun Contact.toIdentity(): MsgLayerIdentity {
        val normalizedPhones = phoneNumbers.map { normalizePhone(it) }.filter { it.isNotBlank() }
        val normalizedEmails = emails.orEmpty().map { it.trim() }.filter { it.isNotBlank() }
        val stableKey = buildString {
            append(id)
            append("|")
            append(name)
            append("|")
            append(normalizedPhones.joinToString(","))
        }
        return MsgLayerIdentity(
            id = "person_${stableHash(stableKey)}",
            type = "person",
            displayName = name,
            phones = normalizedPhones,
            emails = normalizedEmails,
            avatar = photoData,
            labels = listOf("contact"),
            meta = mapOf(
                "source" to "contacts",
                "legacy_contact_id" to id,
                "groups" to groups.orEmpty(),
                "websites" to websites.orEmpty(),
                "note" to note
            )
        )
    }

    private fun resolveParticipantIdentity(
        rawPhone: String,
        identitiesById: MutableMap<String, MsgLayerIdentity>,
        fallbackName: String? = null
    ): String {
        val normalized = normalizePhone(rawPhone)
        identitiesById.values.firstOrNull { identity ->
            identity.type == "person" && identity.phones.any { it == normalized }
        }?.let { return it.id }

        val displayName = fallbackName?.takeIf { it.isNotBlank() } ?: rawPhone.ifBlank { "Unknown" }
        val generated = MsgLayerIdentity(
            id = "person_${stableHash("$displayName|$normalized")}",
            type = "person",
            displayName = displayName,
            phones = listOf(normalized).filter { it.isNotBlank() },
            emails = emptyList(),
            avatar = null,
            labels = listOf("derived"),
            meta = mapOf("source" to "derived")
        )
        identitiesById.putIfAbsent(generated.id, generated)
        return generated.id
    }

    private fun Message.toSmsDirection(): String = when (type) {
        1 -> "inbound"
        2, 4, 5, 6 -> "outbound"
        3 -> "outbound"
        else -> "inbound"
    }

    private fun CallLog.toCallDirection(): String = when (type) {
        2 -> "outbound"
        3 -> "inbound"
        5 -> "missed"
        else -> "missed"
    }

    private fun CallLog.toCallType(): String = when (type) {
        1 -> "missed"
        2 -> "outgoing"
        3 -> "incoming"
        4 -> "voicemail"
        5 -> "rejected"
        else -> "unknown"
    }

    private fun selfIdentityId(deviceId: String): String = "self/$deviceId"

    private fun normalizePhone(raw: String?): String {
        if (raw.isNullOrBlank()) return ""
        val cleaned = raw.filter { it.isDigit() || it == '+' }
        return cleaned.ifBlank { raw.trim() }
    }

    private fun stableHash(input: String): String {
        val digest = MessageDigest.getInstance("SHA-256")
            .digest(input.toByteArray(Charsets.UTF_8))
        return digest.take(8).joinToString("") { "%02x".format(it) }
    }

    private fun Long.toRfc3339(): String {
        return Instant.ofEpochMilli(this)
            .atOffset(ZoneOffset.UTC)
            .format(DateTimeFormatter.ISO_OFFSET_DATE_TIME)
    }
}
