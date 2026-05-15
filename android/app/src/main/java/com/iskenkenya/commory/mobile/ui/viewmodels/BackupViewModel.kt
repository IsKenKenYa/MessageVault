package com.iskenkenya.commory.mobile.ui.viewmodels

import android.content.Context
import android.os.Build
import android.provider.Settings
import androidx.compose.runtime.mutableStateOf
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import com.iskenkenya.commory.mobile.R
import com.iskenkenya.commory.mobile.data.backup.AndroidBackupFileReader
import com.iskenkenya.commory.mobile.data.backup.AndroidBackupFileWriter
import com.iskenkenya.commory.mobile.data.backup.AndroidCallLogReader
import com.iskenkenya.commory.mobile.data.backup.AndroidCallLogWriter
import com.iskenkenya.commory.mobile.data.backup.AndroidContactReader
import com.iskenkenya.commory.mobile.data.backup.AndroidContactWriter
import com.iskenkenya.commory.mobile.data.backup.AndroidSmsReader
import com.iskenkenya.commory.mobile.data.backup.AndroidSmsWriter
import com.iskenkenya.commory.mobile.BuildConfig
import com.iskenkenya.commory.mobile.di.AppContainer
import com.iskenkenya.commory.mobile.remote.CommoryServerClient
import com.iskenkenya.commory.mobile.runtime.AppEnvironmentManager
import com.iskenkenya.commory.mobile.runtime.RuntimeModePolicy
import com.iskenkenya.commory.mobile.ui.model.UiText
import com.iskenkenya.commory.sdk.backup.BackupManager
import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import timber.log.Timber

class BackupViewModel(
    private val backupManager: BackupManager,
    private val deviceIdProvider: () -> String,
    private val environmentManager: AppEnvironmentManager? = null,
    private val serverClient: CommoryServerClient? = null,
    private val dispatcher: CoroutineDispatcher = Dispatchers.IO
) : ViewModel() {

    private val TAG = "BackupViewModel"
    
    private val _permissionsGranted = mutableStateOf(false)
    
    private val _isOperating = mutableStateOf(false)
    
    private val _backupStatus = mutableStateOf<UiText?>(null)
    
    fun getPermissionsGranted(): Boolean = _permissionsGranted.value
    fun isOperating(): Boolean = _isOperating.value
    fun getBackupStatus(): UiText? = _backupStatus.value
    
    fun setPermissionsGranted(granted: Boolean) {
        _permissionsGranted.value = granted
    }
    
    fun startBackup() {
        if (!_permissionsGranted.value) {
            _backupStatus.value = UiText.Resource(R.string.backup_error_permissions)
            return
        }
        
        if (_isOperating.value) {
            Timber.d("[Mobile] DEBUG [Backup] 备份操作已在进行中")
            return
        }
        
        _isOperating.value = true
        _backupStatus.value = UiText.Resource(R.string.backup_status_preparing)
        
        viewModelScope.launch(dispatcher) {
            try {
                val deviceInfo = "${Build.MANUFACTURER} ${Build.MODEL}"
                val environment = environmentManager?.currentSnapshot()
                val result = backupManager.performBackup(
                    hasSmsPermission = true,
                    hasCallLogPermission = true,
                    hasContactsPermission = true,
                    deviceInfo = deviceInfo,
                    deviceId = deviceIdProvider(),
                    appVersion = BuildConfig.VERSION_NAME
                )
                if (result.success) {
                    var remoteUploaded = false
                    val accessToken = environment?.authSession?.accessToken
                    val filePath = result.filePath
                    if (environment != null &&
                        RuntimeModePolicy.canUploadBackup(environment) &&
                        accessToken != null &&
                        filePath != null &&
                        serverClient != null
                    ) {
                        _backupStatus.value = UiText.Resource(R.string.backup_status_uploading)
                        remoteUploaded = serverClient.uploadImport(
                            baseUrl = environment.serverUrl,
                            accessToken = accessToken,
                            file = java.io.File(filePath)
                        ).isSuccess
                    }
                    _backupStatus.value = if (remoteUploaded) {
                        UiText.Resource(
                            R.string.backup_success_remote,
                            listOf(result.smsCount, result.callLogCount)
                        )
                    } else {
                        UiText.Resource(
                            R.string.backup_success_local,
                            listOf(result.smsCount, result.callLogCount)
                        )
                    }
                } else {
                    _backupStatus.value = UiText.Resource(
                        R.string.backup_failed_with_reason,
                        listOf(result.errorMessage.orEmpty())
                    )
                }
            } catch (e: Exception) {
                Timber.e(e, "[Mobile] ERROR [Backup] 备份失败")
                _backupStatus.value = UiText.Resource(
                    R.string.backup_failed_with_reason,
                    listOf(e.message.orEmpty())
                )
            } finally {
                _isOperating.value = false
            }
        }
    }
    
    class Factory(private val context: Context) : ViewModelProvider.Factory {
        @Suppress("UNCHECKED_CAST")
        override fun <T : ViewModel> create(modelClass: Class<T>): T {
            if (modelClass.isAssignableFrom(BackupViewModel::class.java)) {
                val backupManager = BackupManager(
                    smsReader = AndroidSmsReader(context),
                    callLogReader = AndroidCallLogReader(context),
                    contactReader = AndroidContactReader(context),
                    backupFileWriter = AndroidBackupFileWriter(context),
                    backupFileReader = AndroidBackupFileReader(context),
                    smsWriter = AndroidSmsWriter(context),
                    callLogWriter = AndroidCallLogWriter(context),
                    contactWriter = AndroidContactWriter(context)
                )
                return BackupViewModel(
                    backupManager = backupManager,
                    deviceIdProvider = {
                        Settings.Secure.getString(
                            context.contentResolver,
                            Settings.Secure.ANDROID_ID
                        ) ?: "unknown-device"
                    }
                ) as T
            }
            throw IllegalArgumentException("Unknown ViewModel class")
        }
    }

    class ContainerFactory(
        private val context: Context,
        private val container: AppContainer
    ) : ViewModelProvider.Factory {
        @Suppress("UNCHECKED_CAST")
        override fun <T : ViewModel> create(modelClass: Class<T>): T {
            if (modelClass.isAssignableFrom(BackupViewModel::class.java)) {
                return BackupViewModel(
                    backupManager = container.createBackupManager(),
                    deviceIdProvider = {
                        Settings.Secure.getString(
                            context.contentResolver,
                            Settings.Secure.ANDROID_ID
                        ) ?: "unknown-device"
                    },
                    environmentManager = container.environmentManager,
                    serverClient = container.serverClient
                ) as T
            }
            throw IllegalArgumentException("Unknown ViewModel class")
        }
    }
}
