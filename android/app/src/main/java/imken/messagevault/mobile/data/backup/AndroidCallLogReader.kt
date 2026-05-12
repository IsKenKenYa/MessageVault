package imken.messagevault.mobile.data.backup

import android.content.Context
import android.provider.CallLog.Calls
import imken.messagevault.mobile.utils.PhoneNumberUtils
import imken.messagevault.sdk.backup.model.CallLog
import imken.messagevault.sdk.backup.reader.CallLogReader
import timber.log.Timber

class AndroidCallLogReader(private val context: Context) : CallLogReader {

    override suspend fun readCallLogs(): List<CallLog>? {
        val callLogs = mutableListOf<CallLog>()

        try {
            val permissionStatus = context.checkSelfPermission(android.Manifest.permission.READ_CALL_LOG)
            if (permissionStatus != android.content.pm.PackageManager.PERMISSION_GRANTED) {
                Timber.e("[Mobile] ERROR [Backup] 备份通话记录失败: 没有 READ_CALL_LOG 权限")
                return null
            }

            Timber.d("[Mobile] DEBUG [Backup] 开始读取通话记录...")

            val uri = Calls.CONTENT_URI
            val projection = arrayOf(
                Calls._ID,
                Calls.NUMBER,
                Calls.CACHED_NAME,
                Calls.DATE,
                Calls.DURATION,
                Calls.TYPE
            )
            val sortOrder = "${Calls.DATE} DESC"

            val timeRanges = generateTimeRanges()
            var totalCallLogs = 0

            for (range in timeRanges) {
                val selection = "${Calls.DATE} >= ? AND ${Calls.DATE} <= ?"
                val selectionArgs = arrayOf(range.first.toString(), range.second.toString())

                Timber.d("[Mobile] DEBUG [Backup] 查询时间范围: ${range.first} 到 ${range.second}")

                try {
                    context.contentResolver.query(uri, projection, selection, selectionArgs, sortOrder)?.use { cursor ->
                        totalCallLogs += cursor.count

                        val idColumn = cursor.getColumnIndex(Calls._ID)
                        val numberColumn = cursor.getColumnIndex(Calls.NUMBER)
                        val nameColumn = cursor.getColumnIndex(Calls.CACHED_NAME)
                        val dateColumn = cursor.getColumnIndex(Calls.DATE)
                        val durationColumn = cursor.getColumnIndex(Calls.DURATION)
                        val typeColumn = cursor.getColumnIndex(Calls.TYPE)

                        Timber.d("[Mobile] DEBUG [Backup] 批次查询到 ${cursor.count} 条通话记录")

                        while (cursor.moveToNext()) {
                            try {
                                val id = if (idColumn != -1 && !cursor.isNull(idColumn)) cursor.getLong(idColumn) else 0L

                                val originalNumber = if (numberColumn != -1 && !cursor.isNull(numberColumn)) {
                                    cursor.getString(numberColumn)
                                } else {
                                    "unknown"
                                }

                                val normalizedNumber = if (originalNumber != "unknown") {
                                    val normalized = PhoneNumberUtils.normalizePhoneNumber(originalNumber)
                                    if (normalized != originalNumber) {
                                        Timber.d("[Mobile] DEBUG [Backup] 通话记录号码已规范化: $originalNumber -> $normalized")
                                    }
                                    normalized
                                } else {
                                    originalNumber
                                }

                                val name = if (nameColumn != -1 && !cursor.isNull(nameColumn)) {
                                    cursor.getString(nameColumn)
                                } else {
                                    null
                                }

                                val date = if (dateColumn != -1 && !cursor.isNull(dateColumn)) {
                                    cursor.getLong(dateColumn)
                                } else {
                                    System.currentTimeMillis()
                                }

                                val duration = if (durationColumn != -1 && !cursor.isNull(durationColumn)) {
                                    cursor.getInt(durationColumn)
                                } else {
                                    0
                                }

                                val type = if (typeColumn != -1 && !cursor.isNull(typeColumn)) {
                                    cursor.getInt(typeColumn)
                                } else {
                                    Calls.INCOMING_TYPE
                                }

                                val callLog = CallLog(
                                    id = id,
                                    number = normalizedNumber,
                                    contact = name,
                                    date = date,
                                    duration = duration,
                                    type = type
                                )

                                callLogs.add(callLog)
                            } catch (e: Exception) {
                                Timber.e(e, "[Mobile] ERROR [Backup] 处理单条通话记录失败: ${e.message}")
                            }
                        }
                    } ?: run {
                        Timber.e("[Mobile] ERROR [Backup] 查询时间范围 ${range.first} 到 ${range.second} 失败: 返回null")
                    }
                } catch (e: Exception) {
                    Timber.e(e, "[Mobile] ERROR [Backup] 查询时间范围 ${range.first} 到 ${range.second} 失败: ${e.message}")
                }
            }

            Timber.d("[Mobile] DEBUG [Backup] 找到 ${callLogs.size} 条通话记录，总查询数 $totalCallLogs")
            Timber.i("[Mobile] INFO [Backup] 成功读取 ${callLogs.size} 条通话记录")

        } catch (e: Exception) {
            Timber.e(e, "[Mobile] ERROR [Backup] 备份通话记录异常: ${e.message}")
            return callLogs.takeIf { it.isNotEmpty() }
        }

        return callLogs
    }

    private fun generateTimeRanges(ranges: Int = 4): List<Pair<Long, Long>> {
        val result = mutableListOf<Pair<Long, Long>>()
        val endTime = System.currentTimeMillis()
        val startTime = endTime - (365L * 24 * 60 * 60 * 1000)

        val rangeSize = (endTime - startTime) / ranges

        for (i in 0 until ranges) {
            val rangeStart = startTime + (i * rangeSize)
            val rangeEnd = if (i == ranges - 1) endTime else startTime + ((i + 1) * rangeSize - 1)
            result.add(Pair(rangeStart, rangeEnd))
        }

        return result
    }
}
