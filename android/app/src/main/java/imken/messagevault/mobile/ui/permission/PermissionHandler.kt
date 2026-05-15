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
        val BACKUP_PERMISSIONS = PermissionUtils.BACKUP_PERMISSIONS
        val RESTORE_PERMISSIONS = PermissionUtils.RESTORE_PERMISSIONS
    }

    private val requestPermissionLauncher = activity.registerForActivityResult(
        ActivityResultContracts.RequestMultiplePermissions()
    ) { permissions ->
        val allGranted = permissions.entries.all { it.value }
        if (allGranted) {
            Timber.i("[Mobile] INFO [Permissions] 所有请求的权限已授予")
            Toast.makeText(activity, activity.getString(imken.messagevault.mobile.R.string.permissions_granted_toast), Toast.LENGTH_SHORT).show()
        } else {
            val deniedPermissions = permissions.filterValues { !it }.keys
            Timber.w("[Mobile] WARN [Permissions] 部分权限被拒绝: $deniedPermissions")

            if (deniedPermissions.any { activity.shouldShowRequestPermissionRationale(it) }) {
                showPermissionRationaleDialog()
            } else {
                PermissionUtils.openAppSettings(activity)
                Toast.makeText(activity, activity.getString(imken.messagevault.mobile.R.string.permissions_settings_toast), Toast.LENGTH_LONG).show()
            }
        }
    }

    fun checkPermissions(): Boolean = checkBackupPermissions(requestIfMissing = true)

    fun checkBackupPermissions(requestIfMissing: Boolean = true): Boolean {
        val permissionsToCheck = BACKUP_PERMISSIONS

        val allPermissionsGranted = permissionsToCheck.all {
            ContextCompat.checkSelfPermission(activity, it) == PackageManager.PERMISSION_GRANTED
        }

        if (allPermissionsGranted) {
            Timber.i("[Mobile] INFO [Permission] 已有所有必要权限; Context: 启动检查")
        } else {
            Timber.i("[Mobile] INFO [Permission] 请求权限; Context: 启动检查")
            if (requestIfMissing) {
                requestPermissionLauncher.launch(permissionsToCheck)
            }
        }

        return allPermissionsGranted
    }

    fun checkRestorePermissions(requestIfMissing: Boolean = true): Boolean {
        val permissionsToCheck = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            RESTORE_PERMISSIONS
        } else {
            RESTORE_PERMISSIONS + Manifest.permission.WRITE_EXTERNAL_STORAGE
        }
        val allPermissionsGranted = permissionsToCheck.all {
            ContextCompat.checkSelfPermission(activity, it) == PackageManager.PERMISSION_GRANTED
        }
        if (!allPermissionsGranted && requestIfMissing) {
            requestPermissionLauncher.launch(permissionsToCheck)
        }
        return allPermissionsGranted
    }

    private fun showPermissionRationaleDialog() {
        Toast.makeText(activity, activity.getString(imken.messagevault.mobile.R.string.permission_rationale), Toast.LENGTH_LONG).show()
    }
}
