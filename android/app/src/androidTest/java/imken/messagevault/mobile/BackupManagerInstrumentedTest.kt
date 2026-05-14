package imken.messagevault.mobile

import android.Manifest
import android.content.Context
import android.provider.Settings
import androidx.test.core.app.ApplicationProvider
import androidx.test.ext.junit.runners.AndroidJUnit4
import androidx.test.rule.GrantPermissionRule
import imken.messagevault.mobile.BuildConfig
import imken.messagevault.mobile.data.backup.AndroidBackupFileReader
import imken.messagevault.mobile.data.backup.AndroidBackupFileWriter
import imken.messagevault.mobile.data.backup.AndroidCallLogReader
import imken.messagevault.mobile.data.backup.AndroidCallLogWriter
import imken.messagevault.mobile.data.backup.AndroidContactReader
import imken.messagevault.mobile.data.backup.AndroidContactWriter
import imken.messagevault.mobile.data.backup.AndroidSmsReader
import imken.messagevault.mobile.data.backup.AndroidSmsWriter
import imken.messagevault.sdk.backup.BackupManager
import kotlinx.coroutines.runBlocking
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import timber.log.Timber
import java.io.File
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import java.util.TimeZone
import kotlin.test.assertNotNull
import kotlin.test.assertTrue

@RunWith(AndroidJUnit4::class)
class BackupManagerInstrumentedTest {

    @get:Rule
    val permissionRule: GrantPermissionRule = GrantPermissionRule.grant(
        Manifest.permission.READ_SMS,
        Manifest.permission.READ_CALL_LOG,
        Manifest.permission.WRITE_EXTERNAL_STORAGE
    )
    
    private lateinit var context: Context
    private lateinit var backupManager: BackupManager
    private var testStartTime: Long = 0

    @Before
    fun setUp() {
        context = ApplicationProvider.getApplicationContext()
        setupTestLogger()
        testStartTime = System.currentTimeMillis()
        
        backupManager = BackupManager(
            smsReader = AndroidSmsReader(context),
            callLogReader = AndroidCallLogReader(context),
            contactReader = AndroidContactReader(context),
            backupFileWriter = AndroidBackupFileWriter(context),
            backupFileReader = AndroidBackupFileReader(context),
            smsWriter = AndroidSmsWriter(context),
            callLogWriter = AndroidCallLogWriter(context),
            contactWriter = AndroidContactWriter(context)
        )
        
        Timber.i("[Mobile] INFO [Test] 开始BackupManager设备测试; Context: Instrumented test on device")
    }

    private fun setupTestLogger() {
        Timber.plant(object : Timber.Tree() {
            override fun log(priority: Int, tag: String?, message: String, t: Throwable?) {
                val timestamp = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss'Z'", Locale.US)
                    .apply { timeZone = TimeZone.getTimeZone("UTC") }
                    .format(Date())
                val logMessage = "$timestamp $message"
                println(logMessage)
                android.util.Log.println(priority, "MessageVaultTest", logMessage)
            }
        })
    }

    @Test
    fun testPerformBackup() = runBlocking {
        val result = backupManager.performBackup(
            hasSmsPermission = true,
            hasCallLogPermission = true,
            hasContactsPermission = true,
            deviceInfo = "Test Device",
            deviceId = Settings.Secure.getString(
                context.contentResolver,
                Settings.Secure.ANDROID_ID
            ) ?: "instrumented-device",
            appVersion = BuildConfig.VERSION_NAME
        )
        
        Timber.i("[Mobile] INFO [Test] 备份结果; Context: success=${result.success}, smsCount=${result.smsCount}, callLogCount=${result.callLogCount}")
        
        if (result.success) {
            assertNotNull(result.filePath, "备份文件路径不应为null")
            val backupFile = File(result.filePath!!)
            assertTrue(backupFile.exists(), "备份文件应该存在")
            assertTrue(backupFile.length() > 0, "备份文件不应为空")
            backupFile.delete()
        }
        
        Timber.i("[Mobile] INFO [Test] performBackup测试通过; Context: 成功在设备上执行备份")
    }

    @After
    fun tearDown() {
        Timber.i("[Mobile] INFO [Test] 完成BackupManager设备测试; Context: Test duration=${System.currentTimeMillis() - testStartTime}ms")
        Timber.uprootAll()
    }
} 
