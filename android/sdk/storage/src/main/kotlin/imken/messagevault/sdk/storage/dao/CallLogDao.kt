package imken.messagevault.sdk.storage.dao

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import imken.messagevault.sdk.storage.entity.CallLogsEntity

@Dao
interface CallLogDao {

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertCallLogs(callLogs: List<CallLogsEntity>)

    @Query("SELECT * FROM call_logs ORDER BY date DESC")
    suspend fun getAllCallLogs(): List<CallLogsEntity>

    @Query("DELETE FROM call_logs")
    suspend fun deleteAllCallLogs()

    @Query("SELECT COUNT(id) FROM call_logs")
    suspend fun getCallLogsCount(): Int
}
