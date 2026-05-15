package com.iskenkenya.commory.mobile.data.backup

import android.content.ContentResolver
import android.content.ContentValues
import android.content.Context
import android.net.Uri
import android.provider.CallLog
import android.provider.ContactsContract
import com.iskenkenya.commory.mobile.utils.PhoneNumberUtils
import com.iskenkenya.commory.sdk.backup.model.CallLogData
import com.iskenkenya.commory.sdk.backup.writer.CallLogWriter
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.withContext
import timber.log.Timber

class AndroidCallLogWriter(
    private val context: Context
) : CallLogWriter {

    private val contentResolver: ContentResolver = context.contentResolver

    override suspend fun write(callLogs: List<CallLogData>): Int = withContext(Dispatchers.IO) {
        var restoredCount = 0
        val totalCount = callLogs.size
        Timber.i("[Mobile] INFO [Restore] Starting call log restore, total: $totalCount")

        if (!hasCallLogPermissions()) {
            Timber.e("[Mobile] ERROR [Restore] Missing call log permissions")
            return@withContext 0
        }

        val fixedCallLogs = callLogs.map { callLog ->
            if (callLog.number.isBlank()) {
                callLog
            } else {
                val normalizedNumber = PhoneNumberUtils.normalizePhoneNumber(callLog.number)
                if (normalizedNumber != callLog.number) {
                    Timber.d("[Mobile] DEBUG [Restore] Call log number normalized: ${callLog.number} -> $normalizedNumber")
                    callLog.copy(number = normalizedNumber)
                } else {
                    callLog
                }
            }
        }

        val callLogsByNumber = mutableMapOf<String, MutableList<CallLogData>>()

        fixedCallLogs.forEach { callLog ->
            if (callLog.number.isNotBlank()) {
                val normalized = PhoneNumberUtils.normalizePhoneNumber(callLog.number)
                if (!callLogsByNumber.containsKey(normalized)) {
                    callLogsByNumber[normalized] = mutableListOf()
                }
                callLogsByNumber[normalized]!!.add(callLog)
            } else {
                val unknownKey = "unknown_${callLog.id}"
                if (!callLogsByNumber.containsKey(unknownKey)) {
                    callLogsByNumber[unknownKey] = mutableListOf()
                }
                callLogsByNumber[unknownKey]!!.add(callLog)
            }
        }

        Timber.d("[Mobile] DEBUG [Restore] Grouped by normalized number, ${callLogsByNumber.size} unique contacts")

        var processedCount = 0

        for ((_, numberCallLogs) in callLogsByNumber) {
            for (callLog in numberCallLogs) {
                try {
                    val values = ContentValues().apply {
                        put(CallLog.Calls.NUMBER, callLog.number)
                        put(CallLog.Calls.TYPE, callLog.type)
                        put(CallLog.Calls.DATE, callLog.date)
                        put(CallLog.Calls.DURATION, callLog.duration.toLong())
                        put(CallLog.Calls.NEW, 0)

                        if (callLog.number.isNotBlank()) {
                            val contactName = getContactNameFromNumber(callLog.number)
                            if (contactName != null) {
                                put(CallLog.Calls.CACHED_NAME, contactName)
                                Timber.d("[Mobile] DEBUG [Restore] Call log matched contact: ${callLog.number} -> $contactName")
                            }
                        }
                    }

                    val uri = contentResolver.insert(CallLog.Calls.CONTENT_URI, values)
                    if (uri != null) {
                        restoredCount++
                    }

                    processedCount++
                    if (processedCount % 10 == 0) {
                        delay(50)
                    }
                } catch (e: Exception) {
                    Timber.e(e, "[Mobile] ERROR [Restore] Failed to restore call log: number=${callLog.number}, ${e.message}")
                }
            }

            delay(100)
        }

        Timber.i("[Mobile] INFO [Restore] Call log restore complete: success=$restoredCount, total=$totalCount")
        restoredCount
    }

    private fun getContactNameFromNumber(phoneNumber: String): String? {
        if (phoneNumber.isBlank()) return null

        val name = getContactNameByExactNumber(phoneNumber)
        if (name != null) return name

        val variants = PhoneNumberUtils.getPossibleNumberVariants(phoneNumber)
        for (variant in variants) {
            val variantName = getContactNameByExactNumber(variant)
            if (variantName != null) {
                Timber.d("[Mobile] DEBUG [Restore] Matched contact via number variant: $phoneNumber -> $variant -> $variantName")
                return variantName
            }
        }

        return null
    }

    private fun getContactNameByExactNumber(phoneNumber: String): String? {
        val uri = Uri.withAppendedPath(ContactsContract.PhoneLookup.CONTENT_FILTER_URI, Uri.encode(phoneNumber))
        val projection = arrayOf(ContactsContract.PhoneLookup.DISPLAY_NAME)

        try {
            val cursor = contentResolver.query(uri, projection, null, null, null)
            cursor?.use {
                if (it.moveToFirst()) {
                    return it.getString(it.getColumnIndexOrThrow(ContactsContract.PhoneLookup.DISPLAY_NAME))
                }
            }
        } catch (e: Exception) {
            Timber.e(e, "[Mobile] ERROR [Restore] Failed to query contact name: number=$phoneNumber, ${e.message}")
        }

        return null
    }

    private fun hasCallLogPermissions(): Boolean {
        val readPermission = android.Manifest.permission.READ_CALL_LOG
        val writePermission = android.Manifest.permission.WRITE_CALL_LOG

        val readGranted = context.checkSelfPermission(readPermission) == android.content.pm.PackageManager.PERMISSION_GRANTED
        val writeGranted = context.checkSelfPermission(writePermission) == android.content.pm.PackageManager.PERMISSION_GRANTED

        if (!readGranted || !writeGranted) {
            Timber.e("[Mobile] ERROR [Restore] Missing call log permissions: READ_CALL_LOG=$readGranted, WRITE_CALL_LOG=$writeGranted")
            return false
        }

        return true
    }
}
