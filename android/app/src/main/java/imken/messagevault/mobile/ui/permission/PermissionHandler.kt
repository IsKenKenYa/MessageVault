package imken.messagevault.mobile.ui.permission

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.result.contract.ActivityResultContracts
import androidx.core.content.ContextCompat
import imken.messagevault.mobile.utils.PermissionUtils
import timber.log.Timber

class PermissionHandler(private val activity: ComponentActivity) {

    companion object {
        val REQUIRED_PERMISSIONS = arrayOf(
            Manifest.permission.READ_SMS,
            Manifest.permission.READ_CALL_LOG,
            Manifest.permission.READ_CONTACTS,
            Manifest.permission.SEND_SMS,
            Manifest.permission.WRITE_CALL_LOG,
            Manifest.permission.WRITE_CONTACTS,
            Manifest.permission.WRITE_EXTERNAL_STORAGE
        )
    }

    private val requestPermissionLauncher = activity.registerForActivityResult(
        ActivityResultContracts.RequestMultiplePermissions()
    ) { permissions ->
        val allGranted = permissions.entries.all { it.value }
        if (allGranted) {
            Timber.i("[Mobile] INFO [Permissions] 所有请求的权限已授予")
            Toast.makeText(activity, "所有权限已授予，可以继续操作", Toast.LENGTH_SHORT).show()
        } else {
            val deniedPermissions = permissions.filterValues { !it }.keys
            Timber.w("[Mobile] WARN [Permissions] 部分权限被拒绝: $deniedPermissions")

            if (deniedPermissions.any { activity.shouldShowRequestPermissionRationale(it) }) {
                showPermissionRationaleDialog()
            } else {
                PermissionUtils.openAppSettings(activity)
                Toast.makeText(activity, "请在设置中手动授予权限", Toast.LENGTH_LONG).show()
            }
        }
    }

    fun checkPermissions(): Boolean {
        val permissionsToCheck = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            arrayOf(
                Manifest.permission.READ_SMS,
                Manifest.permission.READ_CALL_LOG,
                Manifest.permission.READ_CONTACTS,
                Manifest.permission.SEND_SMS,
                Manifest.permission.WRITE_CALL_LOG,
                Manifest.permission.WRITE_CONTACTS
            )
        } else {
            REQUIRED_PERMISSIONS
        }

        val allPermissionsGranted = permissionsToCheck.all {
            ContextCompat.checkSelfPermission(activity, it) == PackageManager.PERMISSION_GRANTED
        }

        if (allPermissionsGranted) {
            Timber.i("[Mobile] INFO [Permission] 已有所有必要权限; Context: 启动检查")
        } else {
            Timber.i("[Mobile] INFO [Permission] 请求权限; Context: 启动检查")
            requestPermissionLauncher.launch(permissionsToCheck)
        }

        return allPermissionsGranted
    }

    private fun showPermissionRationaleDialog() {
        Toast.makeText(activity, "需要这些权限才能备份和恢复您的短信和通话记录", Toast.LENGTH_LONG).show()
    }
}
