package com.iskenkenya.commory.sdk.backup

import com.iskenkenya.commory.sdk.backup.model.RestoreOptions
import com.iskenkenya.commory.sdk.backup.model.RestoreResult
import com.iskenkenya.commory.sdk.backup.reader.BackupFileReader
import com.iskenkenya.commory.sdk.backup.writer.CallLogWriter
import com.iskenkenya.commory.sdk.backup.writer.ContactWriter
import com.iskenkenya.commory.sdk.backup.writer.SmsWriter
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

class RestoreManager(
    private val backupFileReader: BackupFileReader,
    private val smsWriter: SmsWriter,
    private val callLogWriter: CallLogWriter,
    private val contactWriter: ContactWriter
) {
    suspend fun restore(filePath: String, options: RestoreOptions): RestoreResult =
        withContext(Dispatchers.IO) {
            val backupData = backupFileReader.read(filePath)
                ?: return@withContext RestoreResult(
                    success = false,
                    message = "Unable to parse backup file"
                )

            var restoredSmsCount = 0
            var restoredCallLogsCount = 0
            var restoredContactsCount = 0

            if (options.sms && backupData.messages.isNotEmpty()) {
                try {
                    restoredSmsCount = smsWriter.write(backupData.messages)
                } catch (_: Exception) {
                }
            }

            if (options.callLogs && backupData.callLogs.isNotEmpty()) {
                try {
                    restoredCallLogsCount = callLogWriter.write(backupData.callLogs)
                } catch (_: Exception) {
                }
            }

            if (options.contacts && backupData.contacts.isNotEmpty()) {
                try {
                    restoredContactsCount = contactWriter.write(backupData.contacts)
                } catch (_: Exception) {
                }
            }

            val totalSuccess =
                restoredSmsCount > 0 || restoredCallLogsCount > 0 || restoredContactsCount > 0
            val resultMessage = if (totalSuccess) {
                "Restored $restoredSmsCount SMS, $restoredCallLogsCount call logs, $restoredContactsCount contacts"
            } else {
                "Restore failed: no data was restored"
            }

            RestoreResult(
                success = totalSuccess,
                message = resultMessage,
                restoredSmsCount = restoredSmsCount,
                restoredCallLogsCount = restoredCallLogsCount,
                restoredContactsCount = restoredContactsCount
            )
        }
}
