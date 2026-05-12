package imken.messagevault.mobile.utils

import android.app.AlertDialog
import android.app.role.RoleManager
import android.content.Context
import android.content.Intent
import android.os.Build
import android.provider.Telephony
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.result.contract.ActivityResultContracts
import androidx.lifecycle.ViewModelProvider
import imken.messagevault.mobile.R
import imken.messagevault.mobile.ui.viewmodels.RestoreViewModel
import timber.log.Timber
import java.io.File
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

class DefaultSmsAppHelper(private val activity: ComponentActivity) {

    private val defaultSmsLauncher = activity.registerForActivityResult(
        ActivityResultContracts.StartActivityForResult()
    ) { result ->
        val isSmsAppNow = isDefaultSmsApp()
        if (isSmsAppNow) {
            activity.getSharedPreferences("sms_app_status", Context.MODE_PRIVATE).edit()
                .putBoolean("is_default_sms_app", true)
                .apply()

            val restoreViewModel = ViewModelProvider(activity, RestoreViewModel.Factory(activity))
                .get(RestoreViewModel::class.java)

            restoreViewModel.notifyDefaultSmsAppChanged(true)

            Timber.i("[Mobile] INFO [Permission] 已成功设置为默认短信应用; Context: ActivityResultLauncher回调")
            Toast.makeText(activity, "成功设置为默认短信应用，可以恢复短信", Toast.LENGTH_SHORT).show()

            handleDefaultSmsAppGranted()
        } else {
            activity.getSharedPreferences("sms_app_status", Context.MODE_PRIVATE).edit()
                .putBoolean("is_default_sms_app", false)
                .apply()

            Timber.w("[Mobile] WARN [Permission] 未能设置为默认短信应用; Context: 用户拒绝")
            Toast.makeText(activity, "需要设置为默认短信应用才能恢复短信", Toast.LENGTH_LONG).show()
        }
    }

    fun isDefaultSmsApp(): Boolean {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.KITKAT) {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
                try {
                    val roleManager = activity.getSystemService(Context.ROLE_SERVICE) as? RoleManager
                    if (roleManager != null) {
                        val hasRole = roleManager.isRoleHeld(RoleManager.ROLE_SMS)
                        Timber.d("[Mobile] DEBUG [Permission] RoleManager角色检查结果: $hasRole")
                        if (hasRole) {
                            activity.getSharedPreferences("sms_app_status", Context.MODE_PRIVATE).edit()
                                .putBoolean("is_default_sms_app", true)
                                .apply()
                            return true
                        }
                    }
                } catch (e: Exception) {
                    Timber.e(e, "[Mobile] ERROR [Permission] 检查RoleManager时出错")
                }
            }

            val defaultSmsPackage = Telephony.Sms.getDefaultSmsPackage(activity)
            val isDefault = activity.packageName == defaultSmsPackage

            Timber.d("[Mobile] DEBUG [Permission] 默认短信应用检查: 系统报告默认包名=$defaultSmsPackage, 当前=${activity.packageName}, 是默认=$isDefault")

            activity.getSharedPreferences("sms_app_status", Context.MODE_PRIVATE).edit()
                .putBoolean("is_default_sms_app", isDefault)
                .apply()

            return isDefault
        }
        return true
    }

    fun showDefaultSmsAppDialog() {
        val builder = AlertDialog.Builder(activity)
        builder.setTitle(R.string.permission_required)
            .setMessage("恢复短信需要临时将此应用设置为默认短信应用。\n\n恢复完成后，您可以将其改回原来的应用。\n\n在接下来的系统界面中选择\"是\"，将信驿云储设为默认短信应用，以开始恢复任务。")
            .setPositiveButton(R.string.settings) { _, _ ->
                requestDefaultSmsApp()
            }
            .setNegativeButton(R.string.cancel) { dialog, _ ->
                dialog.dismiss()
            }
            .create()
            .show()
    }

    fun requestDefaultSmsApp() {
        Timber.d("[Mobile] DEBUG [Permission] 开始请求默认短信应用权限，Android版本: ${Build.VERSION.SDK_INT}")
        var requestSent = false

        try {
            if (isDefaultSmsApp()) {
                Timber.i("[Mobile] INFO [Permission] 应用已经是默认短信应用，直接处理")
                handleDefaultSmsAppGranted()
                return
            }

            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
                val roleManager = activity.getSystemService(Context.ROLE_SERVICE) as? RoleManager

                if (roleManager != null) {
                    if (roleManager.isRoleAvailable(RoleManager.ROLE_SMS)) {
                        try {
                            val intent = roleManager.createRequestRoleIntent(RoleManager.ROLE_SMS)
                            defaultSmsLauncher.launch(intent)

                            val message = "使用RoleManager请求SMS角色，Android版本: ${Build.VERSION.SDK_INT}"
                            Timber.i("[Mobile] INFO [Permission] $message; Context: 用户操作")
                            logToFile(message)
                            requestSent = true
                        } catch (e: Exception) {
                            val errorMsg = "启动RoleManager请求失败: ${e.message}"
                            Timber.e(e, "[Mobile] ERROR [Permission] $errorMsg")
                            logToFile(errorMsg)
                        }
                    } else {
                        val warningMsg = "此设备的RoleManager不支持SMS角色 (isRoleAvailable返回false)"
                        Timber.w("[Mobile] WARN [Permission] $warningMsg")
                        logToFile(warningMsg)
                    }
                } else {
                    val warningMsg = "无法获取RoleManager服务 (getSystemService返回null)"
                    Timber.w("[Mobile] WARN [Permission] $warningMsg")
                    logToFile(warningMsg)
                }
            }

            if (!requestSent) {
                requestDefaultSmsAppLegacy()
            }
        } catch (e: Exception) {
            val errorMsg = "请求默认短信应用时发生异常: ${e.message}"
            Timber.e(e, "[Mobile] ERROR [Permission] $errorMsg")
            logToFile(errorMsg)

            if (!requestSent) {
                requestDefaultSmsAppLegacy()
            }
        }
    }

    private fun requestDefaultSmsAppLegacy() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.KITKAT) {
            try {
                val intent = Intent(Telephony.Sms.Intents.ACTION_CHANGE_DEFAULT)
                intent.putExtra(Telephony.Sms.Intents.EXTRA_PACKAGE_NAME, activity.packageName)

                defaultSmsLauncher.launch(intent)

                Timber.d("[Mobile] DEBUG [Permission] 请求成为默认短信应用 (旧API); Context: 用户操作")
                logToFile("使用传统方法请求设置为默认短信应用")
            } catch (e: Exception) {
                Timber.e(e, "[Mobile] ERROR [Permission] 使用传统方法请求默认短信应用失败: ${e.message}")
                logToFile("使用传统方法请求默认短信应用失败: ${e.message}")

                Toast.makeText(
                    activity,
                    "无法自动请求设置默认短信应用，请手动设置",
                    Toast.LENGTH_LONG
                ).show()
            }
        }
    }

    fun showRestoreDefaultSmsAppDialog() {
        val builder = AlertDialog.Builder(activity)
        builder.setTitle("恢复默认短信应用")
            .setMessage("短信恢复已完成。现在您可以将默认短信应用改回原来的应用，也可以稍后再改回。\n\n如果您需要继续恢复其他备份，建议暂时保持本应用为默认短信应用。")
            .setPositiveButton("现在改回") { _, _ ->
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
                    val roleManager = activity.getSystemService(Context.ROLE_SERVICE) as RoleManager
                    val intent = roleManager.createRequestRoleIntent(RoleManager.ROLE_SMS)
                    activity.startActivity(intent)
                } else {
                    val intent = Intent(Telephony.Sms.Intents.ACTION_CHANGE_DEFAULT)
                    intent.putExtra(Telephony.Sms.Intents.EXTRA_PACKAGE_NAME, "")
                    activity.startActivity(intent)
                }
            }
            .setNegativeButton("稍后改回") { dialog, _ ->
                dialog.dismiss()
            }
            .create()
            .show()
    }

    private fun handleDefaultSmsAppGranted() {
        activity.getSharedPreferences("sms_app_status", Context.MODE_PRIVATE).edit()
            .putBoolean("is_default_sms_app", true)
            .apply()

        val restoreViewModel = ViewModelProvider(activity, RestoreViewModel.Factory(activity))
            .get(RestoreViewModel::class.java)

        restoreViewModel.notifyDefaultSmsAppChanged(true)

        val selectedBackupFile = restoreViewModel.selectedBackupFile.value
        if (selectedBackupFile != null) {
            Timber.i("[Mobile] INFO [Restore] 开始恢复备份; Context: 已设置为默认短信应用")
            restoreViewModel.restoreBackupFile(selectedBackupFile)
        } else {
            Timber.w("[Mobile] WARN [Restore] 没有选中备份文件; Context: 已设置为默认短信应用但无备份可恢复")
            Toast.makeText(activity, "没有选中备份文件，请先选择要恢复的备份", Toast.LENGTH_LONG).show()
        }

        logToFile("Fixed SMS role request dialog for Android 16")
    }

    private fun logToFile(message: String) {
        try {
            val logDir = File(activity.filesDir, "logs")
            if (!logDir.exists()) {
                logDir.mkdirs()
            }

            val logFile = File(logDir, "ui-2025-05-13.log")
            val timestamp = SimpleDateFormat("yyyy-MM-dd HH:mm:ss.SSS", Locale.getDefault()).format(Date())
            val logEntry = "$timestamp - $message\n"

            logFile.appendText(logEntry)

            val assetLogDir = File(activity.applicationContext.getExternalFilesDir(null), "logs")
            if (!assetLogDir.exists()) {
                assetLogDir.mkdirs()
            }
            val assetLogFile = File(assetLogDir, "ui-2025-05-13.log")
            assetLogFile.appendText(logEntry)

            if (message.contains("Fixed SMS role request") || message.contains("Fix SMS role request")) {
                assetLogFile.appendText("$timestamp - Fixed SMS role request dialog for Android 16\n")
            }
        } catch (e: Exception) {
            Timber.e(e, "[Mobile] ERROR [Log] 写入日志文件失败: ${e.message}")
        }
    }
}
