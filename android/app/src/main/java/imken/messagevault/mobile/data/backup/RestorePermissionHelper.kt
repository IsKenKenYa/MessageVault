package imken.messagevault.mobile.data.backup

import android.app.Activity
import android.content.Context
import android.content.Intent
import android.net.Uri
import android.os.Build
import android.provider.Settings
import android.provider.Telephony
import android.widget.Toast
import androidx.activity.result.ActivityResultLauncher
import timber.log.Timber

object RestorePermissionHelper {

    fun isDefaultSmsApp(context: Context): Boolean {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.KITKAT) {
            try {
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
                    try {
                        val roleManager = context.getSystemService(Context.ROLE_SERVICE) as? android.app.role.RoleManager
                        if (roleManager != null && roleManager.isRoleHeld(android.app.role.RoleManager.ROLE_SMS)) {
                            return true
                        }
                    } catch (_: Exception) {
                    }
                }
                val defaultSmsPackage = Telephony.Sms.getDefaultSmsPackage(context)
                return defaultSmsPackage == context.packageName
            } catch (_: Exception) {
                return false
            }
        }
        return true
    }

    fun requestDefaultSmsApp(activity: Activity, launcher: ActivityResultLauncher<Intent>) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.KITKAT) {
            try {
                var requestSent = false

                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
                    try {
                        val roleManager = activity.getSystemService(Context.ROLE_SERVICE) as? android.app.role.RoleManager
                        if (roleManager != null && roleManager.isRoleAvailable(android.app.role.RoleManager.ROLE_SMS)) {
                            if (!roleManager.isRoleHeld(android.app.role.RoleManager.ROLE_SMS)) {
                                val roleRequestIntent = roleManager.createRequestRoleIntent(android.app.role.RoleManager.ROLE_SMS)
                                launcher.launch(roleRequestIntent)
                                requestSent = true
                            } else {
                                return
                            }
                        }
                    } catch (_: Exception) {
                    }
                }

                if (!requestSent) {
                    val intent = Intent(Telephony.Sms.Intents.ACTION_CHANGE_DEFAULT)
                    intent.putExtra(Telephony.Sms.Intents.EXTRA_PACKAGE_NAME, activity.packageName)
                    if (intent.resolveActivity(activity.packageManager) != null) {
                        launcher.launch(intent)
                        requestSent = true
                    }
                }

                if (!requestSent) {
                    try {
                        val defaultAppsIntent = Intent(Settings.ACTION_MANAGE_DEFAULT_APPS_SETTINGS)
                        if (defaultAppsIntent.resolveActivity(activity.packageManager) != null) {
                            activity.startActivity(defaultAppsIntent)
                            Toast.makeText(activity, "请在默认应用设置中将本应用设为默认短信应用", Toast.LENGTH_LONG).show()
                            requestSent = true
                        }
                    } catch (_: Exception) {
                    }

                    if (!requestSent) {
                        try {
                            val settingsIntent = Intent(Settings.ACTION_APPLICATION_DETAILS_SETTINGS)
                            settingsIntent.data = Uri.parse("package:" + activity.packageName)
                            activity.startActivity(settingsIntent)
                            Toast.makeText(activity, "请在应用设置中授予短信相关权限，然后设置为默认短信应用", Toast.LENGTH_LONG).show()
                        } catch (_: Exception) {
                            try {
                                val mainSettingsIntent = Intent(Settings.ACTION_SETTINGS)
                                activity.startActivity(mainSettingsIntent)
                                Toast.makeText(activity, "请在系统设置中找到应用管理，将本应用设为默认短信应用", Toast.LENGTH_LONG).show()
                            } catch (_: Exception) {
                                Toast.makeText(activity, "无法自动打开设置，请手动将本应用设为默认短信应用", Toast.LENGTH_LONG).show()
                            }
                        }
                    }
                }
            } catch (e: Exception) {
                Timber.e(e, "[Mobile] ERROR [Restore] Failed to request default SMS app: ${e.message}")
                Toast.makeText(activity, "无法请求默认短信应用权限: ${e.message}", Toast.LENGTH_LONG).show()
            }
        }
    }

    fun checkAndRequestPermissions(activity: Activity, requestCode: Int): Boolean {
        val requiredPermissions = mutableListOf<String>()

        if (!hasSmsPermissions(activity)) {
            requiredPermissions.add(android.Manifest.permission.READ_SMS)
            requiredPermissions.add(android.Manifest.permission.SEND_SMS)
            requiredPermissions.add(android.Manifest.permission.RECEIVE_SMS)
        }

        if (!hasCallLogPermissions(activity)) {
            requiredPermissions.add(android.Manifest.permission.READ_CALL_LOG)
            requiredPermissions.add(android.Manifest.permission.WRITE_CALL_LOG)
        }

        if (!hasContactsPermissions(activity)) {
            requiredPermissions.add(android.Manifest.permission.READ_CONTACTS)
            requiredPermissions.add(android.Manifest.permission.WRITE_CONTACTS)
        }

        if (requiredPermissions.isNotEmpty()) {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
                val specialPermissions = mutableListOf<String>()
                val normalPermissions = mutableListOf<String>()

                for (permission in requiredPermissions) {
                    if (permission == android.Manifest.permission.READ_SMS ||
                        permission == android.Manifest.permission.READ_CALL_LOG ||
                        permission == android.Manifest.permission.WRITE_CALL_LOG) {
                        specialPermissions.add(permission)
                    } else {
                        normalPermissions.add(permission)
                    }
                }

                if (normalPermissions.isNotEmpty()) {
                    activity.requestPermissions(normalPermissions.toTypedArray(), requestCode)
                }

                if (specialPermissions.isNotEmpty()) {
                    try {
                        val intent = Intent(Settings.ACTION_APPLICATION_DETAILS_SETTINGS)
                        intent.data = Uri.parse("package:" + activity.packageName)
                        activity.startActivity(intent)
                    } catch (_: Exception) {
                    }
                }
            } else {
                activity.requestPermissions(requiredPermissions.toTypedArray(), requestCode)
            }
            return false
        }

        return true
    }

    fun hasSmsPermissions(context: Context): Boolean {
        val readGranted = context.checkSelfPermission(android.Manifest.permission.READ_SMS) == android.content.pm.PackageManager.PERMISSION_GRANTED
        val sendGranted = context.checkSelfPermission(android.Manifest.permission.SEND_SMS) == android.content.pm.PackageManager.PERMISSION_GRANTED
        return readGranted && sendGranted
    }

    fun hasCallLogPermissions(context: Context): Boolean {
        val readGranted = context.checkSelfPermission(android.Manifest.permission.READ_CALL_LOG) == android.content.pm.PackageManager.PERMISSION_GRANTED
        val writeGranted = context.checkSelfPermission(android.Manifest.permission.WRITE_CALL_LOG) == android.content.pm.PackageManager.PERMISSION_GRANTED
        return readGranted && writeGranted
    }

    fun hasContactsPermissions(context: Context): Boolean {
        val readGranted = context.checkSelfPermission(android.Manifest.permission.READ_CONTACTS) == android.content.pm.PackageManager.PERMISSION_GRANTED
        val writeGranted = context.checkSelfPermission(android.Manifest.permission.WRITE_CONTACTS) == android.content.pm.PackageManager.PERMISSION_GRANTED
        return readGranted && writeGranted
    }
}
