package imken.messagevault.sdk.backup.writer

import imken.messagevault.sdk.backup.model.CallLogData

interface CallLogWriter {
    suspend fun write(callLogs: List<CallLogData>): Int
}
