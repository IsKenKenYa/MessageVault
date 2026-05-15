package com.iskenkenya.commory.mobile.ui.restore

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.iskenkenya.commory.mobile.data.backup.AndroidBackupFileReader
import com.iskenkenya.commory.mobile.data.backup.AndroidCallLogWriter
import com.iskenkenya.commory.mobile.data.backup.AndroidContactWriter
import com.iskenkenya.commory.mobile.data.backup.AndroidSmsWriter
import com.iskenkenya.commory.mobile.models.BackupFile
import com.iskenkenya.commory.sdk.backup.RestoreManager
import com.iskenkenya.commory.sdk.backup.model.RestoreOptions
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import timber.log.Timber

class RestoreViewModel(
    private val restoreManager: RestoreManager
) : ViewModel() {
    
    private val _restoreState = MutableStateFlow<RestoreState>(RestoreState.Idle)
    val restoreState: StateFlow<RestoreState> = _restoreState
    
    private val _availableBackups = MutableStateFlow<List<BackupFile>>(emptyList())
    val availableBackups: StateFlow<List<BackupFile>> = _availableBackups
    
    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading
    
    fun loadAvailableBackups() {
        viewModelScope.launch {
            _isLoading.value = true
            try {
                _availableBackups.value = emptyList()
                Timber.i("[Mobile] INFO [RestoreViewModel] 加载了 0 个备份文件")
            } catch (e: Exception) {
                Timber.e(e, "[Mobile] ERROR [RestoreViewModel] 加载备份文件失败: ${e.message}")
                _restoreState.value = RestoreState.Error("加载备份文件失败: ${e.message}")
            } finally {
                _isLoading.value = false
            }
        }
    }
    
    fun restoreFromFile(backupFile: BackupFile) {
        viewModelScope.launch {
            try {
                _restoreState.value = RestoreState.InProgress("准备", 0)
                val result = restoreManager.restore(
                    backupFile.filePath,
                    RestoreOptions(sms = true, callLogs = true, contacts = true)
                )
                if (result.success) {
                    _restoreState.value = RestoreState.Success(result.message)
                } else {
                    _restoreState.value = RestoreState.Error(result.message)
                }
            } catch (e: Exception) {
                Timber.e(e, "[Mobile] ERROR [RestoreViewModel] 恢复过程发生异常: ${e.message}")
                _restoreState.value = RestoreState.Error("恢复失败: ${e.message}")
            }
        }
    }
    
    fun cancelRestore() {
        viewModelScope.launch {
            try {
                _restoreState.value = RestoreState.Idle
                Timber.i("[Mobile] INFO [RestoreViewModel] 用户取消了恢复过程")
            } catch (e: Exception) {
                Timber.e(e, "[Mobile] ERROR [RestoreViewModel] 取消恢复过程失败: ${e.message}")
            }
        }
    }
    
    fun resetState() {
        _restoreState.value = RestoreState.Idle
    }
}

sealed class RestoreState {
    object Idle : RestoreState()
    data class InProgress(val operation: String, val progress: Int, val message: String = "") : RestoreState()
    data class Success(val message: String) : RestoreState()
    data class Error(val message: String) : RestoreState()
}
