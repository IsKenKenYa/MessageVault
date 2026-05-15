package com.iskenkenya.commory.sdk.backup.writer

import com.iskenkenya.commory.sdk.backup.model.CallLogData

interface CallLogWriter {
    suspend fun write(callLogs: List<CallLogData>): Int
}
