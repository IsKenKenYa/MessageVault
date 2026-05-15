package com.iskenkenya.commory.mobile.di

import android.content.Context
import com.iskenkenya.commory.mobile.data.backup.AndroidBackupFileReader
import com.iskenkenya.commory.mobile.data.backup.AndroidBackupFileWriter
import com.iskenkenya.commory.mobile.data.backup.AndroidCallLogReader
import com.iskenkenya.commory.mobile.data.backup.AndroidCallLogWriter
import com.iskenkenya.commory.mobile.data.backup.AndroidContactReader
import com.iskenkenya.commory.mobile.data.backup.AndroidContactWriter
import com.iskenkenya.commory.mobile.data.backup.AndroidSmsReader
import com.iskenkenya.commory.mobile.data.backup.AndroidSmsWriter
import com.iskenkenya.commory.mobile.remote.CommoryServerClient
import com.iskenkenya.commory.mobile.runtime.AppEnvironmentManager
import com.iskenkenya.commory.sdk.backup.BackupManager
import com.iskenkenya.commory.sdk.backup.RestoreManager

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
