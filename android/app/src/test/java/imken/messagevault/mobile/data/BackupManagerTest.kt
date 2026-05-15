package imken.messagevault.mobile.data

import android.content.ContentResolver
import android.content.Context
import android.database.MatrixCursor
import android.provider.CallLog.Calls
import android.provider.Telephony
import imken.messagevault.mobile.data.backup.AndroidBackupFileWriter
import imken.messagevault.mobile.data.backup.AndroidCallLogReader
import imken.messagevault.mobile.data.backup.AndroidContactReader
import imken.messagevault.mobile.data.backup.AndroidSmsReader
import imken.messagevault.sdk.backup.BackupManager
import imken.messagevault.sdk.backup.model.BackupData
import imken.messagevault.sdk.backup.model.CallLog
import imken.messagevault.sdk.backup.model.Message
import kotlinx.coroutines.runBlocking
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith
import org.mockito.ArgumentMatchers.anyString
import org.mockito.ArgumentMatchers.any
import org.mockito.ArgumentMatchers.eq
import org.mockito.Mock
import org.mockito.Mockito.`when`
import org.mockito.MockitoAnnotations
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config
import timber.log.Timber
import java.io.StringWriter
import java.text.SimpleDateFormat
import java.util.*
import kotlin.test.assertEquals
import kotlin.test.assertNotNull

@RunWith(RobolectricTestRunner::class)
@Config(manifest = Config.NONE)
class BackupManagerTest {

    @Mock
    private lateinit var mockContext: Context
    
    @Mock
    private lateinit var mockResolver: ContentResolver

    private lateinit var smsReader: AndroidSmsReader
    private lateinit var callLogReader: AndroidCallLogReader
    private lateinit var backupManager: BackupManager
    
    private val logWriter = StringWriter()

    @Before
    fun setUp() {
        MockitoAnnotations.openMocks(this)
        setupTestLogger()
        
        smsReader = AndroidSmsReader(mockContext, mockResolver) { true }
        callLogReader = AndroidCallLogReader(mockContext, mockResolver) { true }
        
        Timber.i("[Mobile] INFO [Test] 开始测试BackupManager; Context: Unit test initialization")
    }
    
    private fun setupTestLogger() {
        Timber.plant(object : Timber.Tree() {
            override fun log(priority: Int, tag: String?, message: String, t: Throwable?) {
                val timestamp = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss'Z'", Locale.US)
                    .apply { timeZone = TimeZone.getTimeZone("UTC") }
                    .format(Date())
                val logMessage = "$timestamp $message"
                logWriter.append(logMessage).append("\n")
                println(logMessage)
            }
        })
    }
    
    @Test
    fun testSmsReaderReadSms() {
        val mockSmsCursor = createMockSmsCursor()
        `when`(mockResolver.query(
            eq(Telephony.Sms.CONTENT_URI),
            any<Array<String>>(),
            any(),
            any(),
            anyString()
        )).thenReturn(mockSmsCursor)
        
        val result = runBlocking { smsReader.readSms() }
        
        assertNotNull(result, "短信列表不应为null")
        assertEquals(2, result!!.size, "应该读取到2条短信记录")
        
        val firstSms = result[0]
        assertEquals("+1234567890", firstSms.address, "第一条短信地址应正确")
        
        Timber.i("[Mobile] INFO [Test] readSms测试通过; Context: 成功验证短信读取功能")
    }
    
    @Test
    fun testCallLogReaderReadCallLogs() {
        val mockCallLogCursor = createMockCallLogCursor()
        `when`(mockResolver.query(
            eq(Calls.CONTENT_URI),
            any<Array<String>>(),
            anyString(),
            any<Array<String>>(),
            anyString()
        )).thenReturn(mockCallLogCursor)
        
        val result = runBlocking { callLogReader.readCallLogs() }
        
        assertNotNull(result, "通话记录列表不应为null")
        assertEquals(2, result!!.size, "应该读取到2条通话记录")
        
        val firstCall = result[0]
        assertEquals("+1234567890", firstCall.number, "第一条通话记录号码应正确")
        
        Timber.i("[Mobile] INFO [Test] readCallLogs测试通过; Context: 成功验证通话记录读取功能")
    }
    
    private fun createMockSmsCursor(): android.database.Cursor {
        val cursor = MatrixCursor(
            arrayOf(
                Telephony.Sms._ID,
                Telephony.Sms.ADDRESS,
                Telephony.Sms.BODY,
                Telephony.Sms.DATE,
                Telephony.Sms.TYPE,
                Telephony.Sms.READ,
                Telephony.Sms.STATUS
            )
        )
        cursor.addRow(arrayOf(1L, "+1234567890", "测试短信内容1", 1714500000000L, 1, 1, 0))
        cursor.addRow(arrayOf(2L, "+0987654321", "测试短信内容2", 1714400000000L, 2, 0, 0))
        return cursor
    }
    
    private fun createMockCallLogCursor(): android.database.Cursor {
        val cursor = MatrixCursor(
            arrayOf(
                Calls._ID,
                Calls.NUMBER,
                Calls.CACHED_NAME,
                Calls.DATE,
                Calls.DURATION,
                Calls.TYPE
            )
        )
        cursor.addRow(arrayOf(1L, "+1234567890", "Test Contact", 1714500000000L, 120, 1))
        cursor.addRow(arrayOf(2L, "+0987654321", null, 1714400000000L, 60, 2))
        return cursor
    }
    
    @org.junit.After
    fun tearDown() {
        Timber.i("[Mobile] INFO [Test] 完成BackupManager测试; Context: Unit test completion")
        Timber.uprootAll()
    }
} 
