package imken.messagevault.sdk.backup.reader

import imken.messagevault.sdk.backup.model.CallLog

interface CallLogReader {
    suspend fun readCallLogs(): List<CallLog>?
}
