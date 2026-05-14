package imken.messagevault.mobile.ui.viewmodels

import android.content.Context
import android.os.Build
import android.provider.Telephony
import androidx.compose.runtime.mutableStateOf
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import imken.messagevault.mobile.api.ApiClient
import imken.messagevault.mobile.config.Config
import imken.messagevault.mobile.data.backup.AndroidBackupFileReader
import imken.messagevault.mobile.data.backup.AndroidCallLogWriter
import imken.messagevault.mobile.data.backup.AndroidContactWriter
import imken.messagevault.mobile.data.backup.AndroidSmsWriter
import imken.messagevault.mobile.data.backup.RestorePermissionHelper
import imken.messagevault.mobile.models.BackupFile
import imken.messagevault.mobile.models.RestoreState
import imken.messagevault.sdk.backup.RestoreManager
import imken.messagevault.sdk.backup.model.BackupData
import imken.messagevault.sdk.backup.model.BackupReadData
import imken.messagevault.sdk.backup.model.CallLog
import imken.messagevault.sdk.backup.model.Contact
import imken.messagevault.sdk.backup.model.Message
import imken.messagevault.sdk.backup.model.RestoreOptions
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import timber.log.Timber
import java.io.File
import java.util.Date

class RestoreViewModel(
    private val context: Context,
    private val restoreManager: RestoreManager,
    private val config: Config,
    private val apiClient: ApiClient
) : ViewModel() {

    private val backupFileReader = AndroidBackupFileReader(context)

    private val _backupFiles = MutableStateFlow<List<BackupFile>>(emptyList())
    val backupFiles: StateFlow<List<BackupFile>> = _backupFiles.asStateFlow()
    
    private val _selectedBackupFile = MutableStateFlow<BackupFile?>(null)
    val selectedBackupFile: StateFlow<BackupFile?> = _selectedBackupFile.asStateFlow()
    
    val restoreStatus = mutableStateOf<String?>(null)
    
    val isOperating = mutableStateOf(false)
    
    val permissionsGranted = mutableStateOf(false)
    
    private val _restoreState = MutableStateFlow<RestoreState>(RestoreState.Idle)
    val restoreState: StateFlow<RestoreState> = _restoreState.asStateFlow()

    private val _restoreProgress = MutableStateFlow(0)
    val restoreProgress: StateFlow<Int> = _restoreProgress.asStateFlow()
    
    val needDefaultSmsApp = mutableStateOf(false)
    
    private var forceDefaultSmsAppStatus: Boolean? = null
    
    val restorePhase = mutableStateOf<String?>(null)
    
    private val TAG = "RestoreViewModel"
    
    init {
        loadBackupFiles()
    }
    
    fun notifyDefaultSmsAppChanged(isDefault: Boolean) {
        forceDefaultSmsAppStatus = isDefault
        Timber.i("[Mobile] INFO [Restore] 强制更新默认短信应用状态: isDefault=$isDefault")
        
        if (isDefault && needDefaultSmsApp.value) {
            needDefaultSmsApp.value = false
            Timber.i("[Mobile] INFO [Restore] 自动重置权限检查对话框，当前被强制设为默认短信应用")
        }
    }
    
    fun loadBackupFiles() {
        if (isOperating.value) {
            Timber.w("[Mobile] WARN [Restore] 已有操作正在进行中，跳过加载备份; Context: 用户请求重复操作")
            return
        }
        
        viewModelScope.launch {
            try {
                Timber.i("[Mobile] INFO [Restore] 开始加载备份文件; Context: 用户打开恢复页面")
                
                val files = getAvailableBackups()
                _backupFiles.value = files
                
                Timber.i("[Mobile] INFO [Restore] 加载备份文件完成; Context: 找到 ${files.size} 个备份文件")
                
                if (files.isNotEmpty() && _selectedBackupFile.value == null) {
                    _selectedBackupFile.value = files[0]
                }
            } catch (e: Exception) {
                Timber.e(e, "[Mobile] ERROR [Restore] 加载备份文件失败; Context: ${e.message}")
                restoreStatus.value = "加载备份文件失败: ${e.message}"
            }
        }
    }
    
    fun selectBackupFile(backupFile: BackupFile) {
        _selectedBackupFile.value = backupFile
        Timber.d("[Mobile] DEBUG [Restore] 选择了备份文件 ${backupFile.fileName}; Context: 用户操作")
    }
    
    fun restoreBackupFile(backupFile: BackupFile) {
        val isDefault = isDefaultSmsApp()
        Timber.i("[Mobile] INFO [Restore] 开始恢复，是否为默认短信应用: $isDefault")
        
        if (isOperating.value) {
            Timber.w("[Mobile] WARN [Restore] 已有操作正在进行; Context: 用户重复点击")
            return
        }
        
        try {
            val backupFilePath = backupFile.filePath
            val jsonFile = java.io.File(backupFilePath)
            if (jsonFile.exists() && jsonFile.canRead()) {
                val jsonText = jsonFile.readText()
                val containsSms = jsonText.contains("\"sms\"") || jsonText.contains("\"messages\"")
                
                Timber.d("[Mobile] DEBUG [Restore] 备份文件解析: 包含短信=${containsSms}, 是默认短信应用=${isDefault}")
                
                if (containsSms && !isDefault) {
                    Timber.w("[Mobile] WARN [Restore] 需要设置为默认短信应用; Context: 用户尝试恢复包含短信的备份")
                    needDefaultSmsApp.value = true
                    return
                }
            } else {
                Timber.e("[Mobile] ERROR [Restore] 备份文件不存在或无法读取; Path: $backupFilePath")
                if (!isDefault) {
                    needDefaultSmsApp.value = true
                    return
                }
            }
        } catch (e: Exception) {
            Timber.e(e, "[Mobile] ERROR [Restore] 检查备份文件是否包含短信时出错; Context: 预检查")
            if (!isDefault) {
                needDefaultSmsApp.value = true
                return
            }
        }
        
        Timber.i("[Mobile] INFO [Restore] 继续恢复过程，忽略WAP推送权限问题")
        
        viewModelScope.launch {
            isOperating.value = true
            _restoreState.value = RestoreState.Preparing
            restoreStatus.value = "正在准备恢复备份..."
            _restoreProgress.value = 0
            
            try {
                Timber.i("%tFT%<tT.%<tLZ [Mobile] INFO [Restore] 开始恢复备份 ${backupFile.fileName}; Context: 用户操作", Date())
                
                restorePhase.value = "正在解析备份文件..."
                _restoreProgress.value = 10
                val backupData = parseBackupFile(backupFile)
                
                if (backupData != null) {
                    val smsCount = backupData.messages?.size ?: 0
                    val callLogsCount = backupData.callLogs?.size ?: 0
                    val contactsCount = backupData.contacts?.size ?: 0
                    
                    Timber.i("[Mobile] INFO [Restore] 备份数据解析完成: SMS=$smsCount, CallLogs=$callLogsCount, Contacts=$contactsCount")
                    
                    _restoreProgress.value = 20
                    restorePhase.value = "正在恢复短信、通话记录和联系人..."
                    _restoreState.value = RestoreState.InProgress(20, "正在恢复短信、通话记录和联系人...")
                    
                    val hasSms = !backupData.messages.isNullOrEmpty()
                    val hasCallLogs = !backupData.callLogs.isNullOrEmpty()
                    val hasContacts = !backupData.contacts.isNullOrEmpty()
                    
                    if (hasSms && !RestorePermissionHelper.isDefaultSmsApp(context)) {
                        withContext(Dispatchers.Main) {
                            isOperating.value = false
                            _restoreProgress.value = 0
                            _restoreState.value = RestoreState.Error("恢复短信需要将此应用设为默认短信应用")
                            restoreStatus.value = "恢复失败: 恢复短信需要将此应用设为默认短信应用"
                            restorePhase.value = null
                        }
                        return@launch
                    }
                    
                    val result = restoreManager.restore(
                        backupFile.filePath,
                        RestoreOptions(sms = hasSms, callLogs = hasCallLogs, contacts = hasContacts)
                    )
                    
                    _restoreProgress.value = 80
                    
                    withContext(Dispatchers.Main) {
                        isOperating.value = false
                        _restoreProgress.value = 100
                        
                        if (result.success) {
                            _restoreState.value = RestoreState.Completed(true, result.message)
                            restoreStatus.value = result.message
                            restorePhase.value = null
                            Timber.i("[Mobile] INFO [Restore] 恢复备份成功; Context: ${result.message}")
                        } else {
                            _restoreState.value = RestoreState.Completed(false, result.message)
                            restoreStatus.value = result.message
                            restorePhase.value = null
                            Timber.e("[Mobile] ERROR [Restore] 恢复备份失败; Context: ${result.message}")
                        }
                    }
                } else {
                    withContext(Dispatchers.Main) {
                        isOperating.value = false
                        _restoreProgress.value = 0
                        _restoreState.value = RestoreState.Error("无法解析备份文件")
                        restoreStatus.value = "恢复失败: 无法解析备份文件"
                        restorePhase.value = null
                        Timber.e("[Mobile] ERROR [Restore] 无法解析备份文件; Context: 文件格式可能不兼容")
                    }
                }
            } catch (e: Exception) {
                withContext(Dispatchers.Main) {
                    isOperating.value = false
                    _restoreProgress.value = 0
                    _restoreState.value = RestoreState.Error(e.message ?: "未知错误")
                    restoreStatus.value = "恢复失败: ${e.message}"
                    restorePhase.value = null
                    Timber.e(e, "[Mobile] ERROR [Restore] 恢复过程异常; Context: ${e.message}")
                }
            }
        }
    }
    
    private fun isDefaultSmsApp(): Boolean {
        if (forceDefaultSmsAppStatus != null) {
            val forcedStatus = forceDefaultSmsAppStatus!!
            Timber.d("[Mobile] DEBUG [Restore] 使用强制设置的默认短信应用状态: $forcedStatus")
            return forcedStatus
        }
        
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.KITKAT) {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
                try {
                    val roleManager = context.getSystemService(Context.ROLE_SERVICE) as? android.app.role.RoleManager
                    if (roleManager != null) {
                        val hasRole = roleManager.isRoleHeld(android.app.role.RoleManager.ROLE_SMS)
                        Timber.d("[Mobile] DEBUG [Restore] RoleManager检查结果: $hasRole")
                        if (hasRole) {
                            return true
                        }
                    }
                } catch (e: Exception) {
                    Timber.e(e, "[Mobile] ERROR [Restore] 使用RoleManager检查权限失败")
                }
            }
            
            val defaultSmsPackage = Telephony.Sms.getDefaultSmsPackage(context)
            val isDefault = context.packageName == defaultSmsPackage
            
            Timber.d("[Mobile] DEBUG [Restore] 默认短信应用检查: 当前=${context.packageName}, 系统默认=$defaultSmsPackage, 是默认=${isDefault}")
            
            return isDefault
        }
        return true
    }
    
    private suspend fun getAvailableBackups(): List<BackupFile> = withContext(Dispatchers.IO) {
        val backupFiles = mutableListOf<BackupFile>()
        try {
            val backupDir = File(context.getExternalFilesDir(null), config.getBackupDirectoryName())
            if (!backupDir.exists()) return@withContext emptyList<BackupFile>()

            val files = backupDir.listFiles { file ->
                file.isFile && file.name.endsWith(".json", ignoreCase = true)
            }

            if (files != null) {
                val validFiles = files.filter { validateBackupFile(it) }
                backupFiles.addAll(validFiles.map { file ->
                    val deviceId = android.provider.Settings.Secure.getString(
                        context.contentResolver, android.provider.Settings.Secure.ANDROID_ID
                    ) ?: "unknown"

                    val backupData = backupFileReader.read(file.absolutePath)

                    BackupFile(
                        filePath = file.absolutePath,
                        fileName = file.name,
                        fileSize = file.length(),
                        creationDate = Date(file.lastModified()),
                        deviceName = backupData?.deviceInfo ?: deviceId,
                        smsCount = backupData?.messages?.size ?: 0,
                        callLogsCount = backupData?.callLogs?.size ?: 0,
                        version = imken.messagevault.mobile.BuildConfig.VERSION_NAME
                    )
                }.sortedByDescending { it.creationDate.time })
            }
        } catch (e: Exception) {
            Timber.e(e, "[Mobile] ERROR [Restore] Failed to get backup file list: ${e.message}")
        }
        backupFiles
    }

    private suspend fun parseBackupFile(backupFile: BackupFile): BackupData? = withContext(Dispatchers.IO) {
        backupFileReader.read(backupFile.filePath)?.toLegacyBackupData()
    }

    private fun validateBackupFile(file: File): Boolean {
        if (!file.exists() || !file.isFile || !file.canRead()) return false
        return runCatching { kotlinx.coroutines.runBlocking { backupFileReader.read(file.absolutePath) != null } }
            .getOrDefault(false)
    }

    private fun BackupReadData.toLegacyBackupData(): BackupData {
        return BackupData(
            messages = messages.map { message ->
                Message(
                    id = message.id,
                    address = message.address,
                    body = message.body,
                    date = message.date,
                    type = message.type,
                    readState = message.readState,
                    messageStatus = message.messageStatus,
                    threadId = message.threadId
                )
            },
            callLogs = callLogs.map { callLog ->
                CallLog(
                    id = callLog.id,
                    number = callLog.number,
                    type = callLog.type,
                    date = callLog.date,
                    duration = callLog.duration,
                    contact = callLog.contact
                )
            },
            contacts = contacts.map { contact ->
                Contact(
                    id = contact.id,
                    name = contact.name,
                    phoneNumbers = contact.phoneNumbers.toMutableList(),
                    emails = contact.emails
                )
            },
            timestamp = timestamp,
            deviceInfo = deviceInfo
        )
    }
    
    fun setPermissionsGranted(granted: Boolean) {
        permissionsGranted.value = granted
        if (granted) {
            loadBackupFiles()
        }
    }
    
    fun resetNeedDefaultSmsApp() {
        needDefaultSmsApp.value = false
        val isDefault = isDefaultSmsApp()
        Timber.i("[Mobile] INFO [Restore] 重置权限检查对话框，当前是否为默认短信应用: $isDefault")
    }
    
    fun checkAndUpdateDefaultSmsAppStatus(): Boolean {
        forceDefaultSmsAppStatus = null
        val isDefault = isDefaultSmsApp()
        if (isDefault) {
            forceDefaultSmsAppStatus = true
            Timber.i("[Mobile] INFO [Restore] 检测到应用已是默认短信应用，更新状态")
        }
        return isDefault
    }
    
    class Factory(private val context: Context) : ViewModelProvider.Factory {
        @Suppress("UNCHECKED_CAST")
        override fun <T : ViewModel> create(modelClass: Class<T>): T {
            if (modelClass.isAssignableFrom(RestoreViewModel::class.java)) {
                val restoreManager = RestoreManager(
                    backupFileReader = AndroidBackupFileReader(context),
                    smsWriter = AndroidSmsWriter(context),
                    callLogWriter = AndroidCallLogWriter(context),
                    contactWriter = AndroidContactWriter(context)
                )
                val config = Config.getInstance(context)
                val apiClient = ApiClient(config)
                
                return RestoreViewModel(context, restoreManager, config, apiClient) as T
            }
            throw IllegalArgumentException("Unknown ViewModel class")
        }
    }

    fun logError(message: String, exception: Exception? = null) {
        Timber.e(exception, message)
    }
} 
