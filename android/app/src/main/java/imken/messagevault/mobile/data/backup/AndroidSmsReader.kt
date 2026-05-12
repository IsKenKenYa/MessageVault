package imken.messagevault.mobile.data.backup

import android.content.Context
import android.provider.Telephony
import imken.messagevault.mobile.utils.PhoneNumberUtils
import imken.messagevault.sdk.backup.model.Message
import imken.messagevault.sdk.backup.reader.SmsReader
import timber.log.Timber

class AndroidSmsReader(private val context: Context) : SmsReader {

    override suspend fun readSms(): List<Message>? {
        val messages = mutableListOf<Message>()

        try {
            val permissionStatus = context.checkSelfPermission(android.Manifest.permission.READ_SMS)
            if (permissionStatus != android.content.pm.PackageManager.PERMISSION_GRANTED) {
                Timber.e("[Mobile] ERROR [Backup] 备份短信失败: 没有 READ_SMS 权限")
                return null
            }

            Timber.d("[Mobile] DEBUG [Backup] 开始读取短信...")

            val uri = Telephony.Sms.CONTENT_URI
            val projection = arrayOf(
                Telephony.Sms._ID,
                Telephony.Sms.ADDRESS,
                Telephony.Sms.BODY,
                Telephony.Sms.DATE,
                Telephony.Sms.TYPE,
                Telephony.Sms.READ,
                Telephony.Sms.STATUS
            )
            val sortOrder = "${Telephony.Sms.DATE} DESC"

            context.contentResolver.query(uri, projection, null, null, sortOrder)?.use { cursor ->
                Timber.d("[Mobile] DEBUG [Backup] 找到 ${cursor.count} 条短信")

                val idColumn = cursor.getColumnIndex(Telephony.Sms._ID)
                val addressColumn = cursor.getColumnIndex(Telephony.Sms.ADDRESS)
                val bodyColumn = cursor.getColumnIndex(Telephony.Sms.BODY)
                val dateColumn = cursor.getColumnIndex(Telephony.Sms.DATE)
                val typeColumn = cursor.getColumnIndex(Telephony.Sms.TYPE)
                val readColumn = cursor.getColumnIndex(Telephony.Sms.READ)
                val statusColumn = cursor.getColumnIndex(Telephony.Sms.STATUS)

                while (cursor.moveToNext()) {
                    val id = if (idColumn != -1) cursor.getLong(idColumn) else 0

                    val originalAddress = if (addressColumn != -1) cursor.getString(addressColumn) else ""

                    val normalizedAddress = if (originalAddress.isNotBlank()) {
                        val normalized = PhoneNumberUtils.normalizePhoneNumber(originalAddress)
                        if (normalized != originalAddress) {
                            Timber.d("[Mobile] DEBUG [Backup] 短信地址已规范化: $originalAddress -> $normalized")
                        }
                        normalized
                    } else {
                        originalAddress
                    }

                    val body = if (bodyColumn != -1) cursor.getString(bodyColumn) else ""
                    val date = if (dateColumn != -1) cursor.getLong(dateColumn) else 0
                    val type = if (typeColumn != -1) cursor.getInt(typeColumn) else 0
                    val read = if (readColumn != -1) cursor.getInt(readColumn) else 0
                    val status = if (statusColumn != -1) cursor.getInt(statusColumn) else 0

                    val message = Message(
                        id = id,
                        address = normalizedAddress,
                        body = body,
                        date = date,
                        type = type,
                        readState = read,
                        messageStatus = status
                    )

                    messages.add(message)
                }
            } ?: run {
                Timber.e("[Mobile] ERROR [Backup] 备份短信失败: 无法查询短信内容提供者")
                return null
            }

            Timber.i("[Mobile] INFO [Backup] 成功读取 ${messages.size} 条短信")

        } catch (e: Exception) {
            Timber.e(e, "[Mobile] ERROR [Backup] 备份短信异常: ${e.message}")
            return null
        }

        return messages
    }
}
