package com.iskenkenya.commory.sdk.backup.reader

import com.iskenkenya.commory.sdk.backup.model.CallLog

interface CallLogReader {
    suspend fun readCallLogs(): List<CallLog>?
}
