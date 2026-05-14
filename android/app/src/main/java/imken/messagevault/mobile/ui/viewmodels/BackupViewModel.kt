package imken.messagevault.mobile.ui.viewmodels

import android.content.Context
import android.os.Build
import android.provider.Settings
import androidx.compose.runtime.State
import androidx.compose.runtime.mutableStateOf
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import imken.messagevault.mobile.data.backup.AndroidBackupFileReader
import imken.messagevault.mobile.data.backup.AndroidBackupFileWriter
import imken.messagevault.mobile.data.backup.AndroidCallLogReader
import imken.messagevault.mobile.data.backup.AndroidCallLogWriter
import imken.messagevault.mobile.data.backup.AndroidContactReader
import imken.messagevault.mobile.data.backup.AndroidContactWriter
import imken.messagevault.mobile.data.backup.AndroidSmsReader
import imken.messagevault.mobile.data.backup.AndroidSmsWriter
import imken.messagevault.mobile.BuildConfig
import imken.messagevault.sdk.backup.BackupManager
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
    private val dispatcher: CoroutineDispatcher = Dispatchers.IO
) : ViewModel() {

    private val TAG = "BackupViewModel"
    
    private val _permissionsGranted = mutableStateOf(false)
    
    private val _isOperating = mutableStateOf(false)
    
    private val _backupStatus = mutableStateOf<String?>(null)
    
    fun getPermissionsGranted(): Boolean = _permissionsGranted.value
    fun isOperating(): Boolean = _isOperating.value
    fun getBackupStatus(): String? = _backupStatus.value
    
    fun setPermissionsGranted(granted: Boolean) {
        _permissionsGranted.value = granted
    }
    
    fun startBackup() {
        if (!_permissionsGranted.value) {
            _backupStatus.value = "权限不足，无法执行备份"
            return
        }
        
        if (_isOperating.value) {
            Timber.d("[Mobile] DEBUG [Backup] 备份操作已在进行中")
            return
        }
        
        _isOperating.value = true
        _backupStatus.value = "正在准备备份..."
        
        viewModelScope.launch(dispatcher) {
            try {
                val deviceInfo = "${Build.MANUFACTURER} ${Build.MODEL}"
                val result = backupManager.performBackup(
                    hasSmsPermission = true,
                    hasCallLogPermission = true,
                    hasContactsPermission = true,
                    deviceInfo = deviceInfo,
                    deviceId = deviceIdProvider(),
                    appVersion = BuildConfig.VERSION_NAME
                )
                if (result.success) {
                    _backupStatus.value = "备份完成: ${result.smsCount} 条短信, ${result.callLogCount} 条通话记录"
                } else {
                    _backupStatus.value = "备份失败: ${result.errorMessage}"
                }
            } catch (e: Exception) {
                Timber.e(e, "[Mobile] ERROR [Backup] 备份失败")
                _backupStatus.value = "备份失败: ${e.message}"
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
}
