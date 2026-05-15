package com.iskenkenya.commory.mobile

import android.content.Context
import android.content.res.Configuration
import android.os.Bundle
import android.util.Log
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.*
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.iskenkenya.commory.mobile.config.Config
import com.iskenkenya.commory.mobile.di.AppContainer
import com.iskenkenya.commory.mobile.runtime.AppEnvironmentManager
import com.iskenkenya.commory.mobile.runtime.LocaleResolver
import com.iskenkenya.commory.mobile.ui.navigation.CommoryAppWithNavigation
import com.iskenkenya.commory.mobile.ui.navigation.NavigationItem
import com.iskenkenya.commory.mobile.ui.navigation.navigationItems
import com.iskenkenya.commory.mobile.ui.permission.PermissionHandler
import com.iskenkenya.commory.mobile.ui.theme.CommoryTheme
import com.iskenkenya.commory.mobile.ui.viewmodels.AppViewModel
import com.iskenkenya.commory.mobile.ui.viewmodels.BackupViewModel
import com.iskenkenya.commory.mobile.ui.viewmodels.RestoreViewModel
import com.iskenkenya.commory.mobile.utils.DefaultSmsAppHelper
import com.iskenkenya.commory.mobile.utils.PermissionUtils
import timber.log.Timber
import java.util.Locale

class MainActivity : ComponentActivity() {

    companion object {
        private const val TAG = "MainActivity"
    }

    private lateinit var config: Config
    private lateinit var appContainer: AppContainer
    private lateinit var permissionHandler: PermissionHandler
    private lateinit var defaultSmsAppHelper: DefaultSmsAppHelper

    private var initialPermissionsChecked = false

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        if (Timber.forest().isEmpty()) {
            Timber.plant(object : Timber.DebugTree() {
                override fun log(priority: Int, tag: String?, message: String, t: Throwable?) {
                    Log.println(priority, "Commory", message)
                    super.log(priority, tag, message, t)
                }
            })
        }

        Timber.d("[Mobile] DEBUG 日志系统初始化完成")
        Log.d("Commory", "主活动创建 - 直接Log测试")

        config = Config.getInstance(this)
        appContainer = AppContainer(applicationContext)
        permissionHandler = PermissionHandler(this)
        defaultSmsAppHelper = DefaultSmsAppHelper(this)

        applyLanguage()

        setContent {
            CommoryTheme {
                val appViewModel: AppViewModel = viewModel(
                    factory = AppViewModel.Factory(appContainer)
                )

                val backupViewModel: BackupViewModel = viewModel(
                    factory = BackupViewModel.ContainerFactory(this, appContainer)
                )

                val restoreViewModel: RestoreViewModel = viewModel(
                    factory = RestoreViewModel.Factory(this)
                )

                if (!initialPermissionsChecked) {
                    val permissionsGranted = permissionHandler.checkBackupPermissions()
                    backupViewModel.setPermissionsGranted(permissionsGranted)
                    restoreViewModel.setPermissionsGranted(permissionsGranted)
                    initialPermissionsChecked = true
                }

                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    CommoryAppWithNavigation(
                        appViewModel = appViewModel,
                        backupViewModel = backupViewModel,
                        restoreViewModel = restoreViewModel,
                        navigationItems = navigationItems
                    )
                }
            }
        }

        Timber.i("[Mobile] INFO [UI] 实现MD3 UI与底部导航; Context: 应用启动")
    }

    private fun applyLanguage() {
        val environment = AppEnvironmentManager(this).currentSnapshot()
        val locale = LocaleResolver.resolveLocale(environment)
        Locale.setDefault(locale)

        val configuration = Configuration(resources.configuration)
        configuration.setLocale(locale)
        createConfigurationContext(configuration)
    }

    override fun attachBaseContext(newBase: Context) {
        val environment = AppEnvironmentManager(newBase).currentSnapshot()
        super.attachBaseContext(LocaleResolver.wrapContext(newBase, environment))
    }

    override fun onResume() {
        super.onResume()
        if (PermissionUtils.checkAllRequiredPermissions(this)) {
            Timber.d("[Mobile] DEBUG [Permissions] 所有需要的权限已授予")
        }
    }

    fun requestDefaultSmsApp() {
        defaultSmsAppHelper.requestDefaultSmsApp()
    }

    fun isDefaultSmsApp(): Boolean {
        return defaultSmsAppHelper.isDefaultSmsApp()
    }

    fun showDefaultSmsAppDialog() {
        defaultSmsAppHelper.showDefaultSmsAppDialog()
    }

    fun showRestoreDefaultSmsAppDialog() {
        defaultSmsAppHelper.showRestoreDefaultSmsAppDialog()
    }
}

@Preview(showBackground = true)
@Composable
fun CommoryAppPreview() {
    CommoryTheme {
        Surface(
            modifier = Modifier.fillMaxSize(),
            color = MaterialTheme.colorScheme.background
        ) {
            Column(
                modifier = Modifier.padding(16.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.Center
            ) {
                androidx.compose.material3.Text(
                    text = stringResource(id = R.string.app_title),
                    style = MaterialTheme.typography.headlineLarge
                )

                Spacer(modifier = Modifier.height(16.dp))

                androidx.compose.material3.Text(
                    text = stringResource(id = R.string.app_description),
                    style = MaterialTheme.typography.bodyLarge
                )
            }
        }
    }
}
