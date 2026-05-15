package imken.messagevault.mobile.di

import android.content.Context
import imken.messagevault.mobile.data.backup.AndroidBackupFileReader
import imken.messagevault.mobile.data.backup.AndroidBackupFileWriter
import imken.messagevault.mobile.data.backup.AndroidCallLogReader
import imken.messagevault.mobile.data.backup.AndroidCallLogWriter
import imken.messagevault.mobile.data.backup.AndroidContactReader
import imken.messagevault.mobile.data.backup.AndroidContactWriter
import imken.messagevault.mobile.data.backup.AndroidSmsReader
import imken.messagevault.mobile.data.backup.AndroidSmsWriter
import imken.messagevault.mobile.remote.CommoryServerClient
import imken.messagevault.mobile.runtime.AppEnvironmentManager
import imken.messagevault.sdk.backup.BackupManager
import imken.messagevault.sdk.backup.RestoreManager

class AppContainer(private val context: Context) {
    val environmentManager: AppEnvironmentManager by lazy { AppEnvironmentManager(context) }
    val serverClient: CommoryServerClient by lazy { CommoryServerClient(context, environmentManager) }

    fun createBackupManager(): BackupManager = BackupManager(
        smsReader = AndroidSmsReader(context),
        callLogReader = AndroidCallLogReader(context),
        contactReader = AndroidContactReader(context),
        backupFileWriter = AndroidBackupFileWriter(context),
        backupFileReader = AndroidBackupFileReader(context),
        smsWriter = AndroidSmsWriter(context),
        callLogWriter = AndroidCallLogWriter(context),
        contactWriter = AndroidContactWriter(context)
    )

    fun createRestoreManager(): RestoreManager = RestoreManager(
        backupFileReader = AndroidBackupFileReader(context),
        smsWriter = AndroidSmsWriter(context),
        callLogWriter = AndroidCallLogWriter(context),
        contactWriter = AndroidContactWriter(context)
    )
}
