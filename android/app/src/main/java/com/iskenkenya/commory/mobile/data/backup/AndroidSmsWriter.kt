package com.iskenkenya.commory.mobile.data.backup

import android.content.ContentValues
import android.content.Context
import android.provider.Telephony
import com.iskenkenya.commory.sdk.backup.model.SmsData
import com.iskenkenya.commory.sdk.backup.writer.SmsWriter
import kotlinx.coroutines.delay
import timber.log.Timber

class AndroidSmsWriter(private val context: Context) : SmsWriter {

    override suspend fun write(messages: List<SmsData>): Int {
        var restoredCount = 0
        val totalCount = messages.size
        Timber.i("[Mobile] INFO [Restore] 开始恢复短信，总数: $totalCount")

        val fixedMessages = validateAndFixMessageAddresses(messages)
        val messagesByContact = fixedMessages.groupBy { it.address }

        var processedCount = 0

        for ((address, contactMessages) in messagesByContact) {
            Timber.d("[Mobile] DEBUG [Restore] 处理联系人消息: 联系人=$address, 消息数量=${contactMessages.size}")

            val sortedMessages = contactMessages.sortedBy { it.date }

            for (smsMessage in sortedMessages) {
                try {
                    val values = ContentValues().apply {
                        put(Telephony.Sms.ADDRESS, smsMessage.address)
                        put(Telephony.Sms.BODY, smsMessage.body)
                        put(Telephony.Sms.DATE, smsMessage.date)
                        put(Telephony.Sms.TYPE, smsMessage.type)
                        put(Telephony.Sms.READ, smsMessage.readState ?: 0)
                        put(Telephony.Sms.STATUS, smsMessage.messageStatus ?: 0)

                        if (smsMessage.threadId > 0) {
                            put(Telephony.Sms.THREAD_ID, smsMessage.threadId)
                        }
                    }

                    val uri = context.contentResolver.insert(Telephony.Sms.CONTENT_URI, values)
                    if (uri != null) {
                        restoredCount++
                    }

                    processedCount++

                    if (processedCount % 10 == 0) {
                        delay(50)
                    }
                } catch (e: Exception) {
                    Timber.e(e, "[Mobile] ERROR [Restore] 恢复短信失败: 地址=${smsMessage.address}, ${e.message}")
                }
            }

            delay(100)
        }

        Timber.i("[Mobile] INFO [Restore] 短信恢复完成: 成功=$restoredCount, 总数=$totalCount")
        return restoredCount
    }

    private fun validateAndFixMessageAddresses(messages: List<SmsData>): List<SmsData> {
        Timber.d("[Mobile] DEBUG [Restore] 开始验证和修复短信地址，数量: ${messages.size}")

        val emptyAddressCount = messages.count { it.address.isNullOrBlank() }
        if (emptyAddressCount > 0) {
            Timber.w("[Mobile] WARN [Restore] 发现 $emptyAddressCount 条短信缺少有效地址，将进行修复")
        }

        return messages.map { originalMessage ->
            if (originalMessage.address.isNullOrBlank()) {
                SmsData(
                    id = originalMessage.id,
                    address = "unknown_${originalMessage.id}",
                    body = originalMessage.body,
                    date = originalMessage.date,
                    type = originalMessage.type,
                    readState = originalMessage.readState,
                    messageStatus = originalMessage.messageStatus,
                    threadId = originalMessage.threadId
                )
            } else {
                val normalizedAddress = com.iskenkenya.commory.mobile.utils.PhoneNumberUtils.normalizePhoneNumber(originalMessage.address)

                if (normalizedAddress != originalMessage.address) {
                    Timber.d("[Mobile] DEBUG [Restore] 标准化地址: ${originalMessage.address} -> $normalizedAddress")

                    SmsData(
                        id = originalMessage.id,
                        address = normalizedAddress,
                        body = originalMessage.body,
                        date = originalMessage.date,
                        type = originalMessage.type,
                        readState = originalMessage.readState,
                        messageStatus = originalMessage.messageStatus,
                        threadId = originalMessage.threadId
                    )
                } else {
                    originalMessage
                }
            }
        }
    }
}
