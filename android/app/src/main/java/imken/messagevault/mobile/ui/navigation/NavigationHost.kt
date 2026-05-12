package imken.messagevault.mobile.ui.navigation

import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Backup
import androidx.compose.material.icons.filled.MoreHoriz
import androidx.compose.material.icons.filled.Preview
import androidx.compose.material.icons.filled.Restore
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.stringResource
import androidx.navigation.NavDestination.Companion.hierarchy
import androidx.navigation.NavGraph.Companion.findStartDestination
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import imken.messagevault.mobile.R
import imken.messagevault.mobile.ui.screens.BackupScreen
import imken.messagevault.mobile.ui.screens.MoreScreen
import imken.messagevault.mobile.ui.screens.PreviewScreen
import imken.messagevault.mobile.ui.screens.RestoreScreen
import imken.messagevault.mobile.ui.viewmodels.BackupViewModel
import imken.messagevault.mobile.ui.viewmodels.RestoreViewModel

sealed class NavigationItem(val route: String, val iconData: ImageVector, val labelResId: Int) {
    object Backup : NavigationItem("backup", Icons.Filled.Backup, R.string.backup_tab)
    object Restore : NavigationItem("restore", Icons.Filled.Restore, R.string.restore_tab)
    object Preview : NavigationItem("preview", Icons.Filled.Preview, R.string.preview_tab)
    object More : NavigationItem("more", Icons.Filled.MoreHoriz, R.string.more_tab)
}

val navigationItems = listOf(
    NavigationItem.Backup,
    NavigationItem.Restore,
    NavigationItem.Preview,
    NavigationItem.More
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MessageVaultAppWithNavigation(
    backupViewModel: BackupViewModel,
    restoreViewModel: RestoreViewModel,
    navigationItems: List<NavigationItem>
) {
    val navController = rememberNavController()

    Scaffold(
        topBar = {
            CenterAlignedTopAppBar(
                title = {
                    Text(
                        text = stringResource(id = R.string.app_title),
                        style = MaterialTheme.typography.titleLarge
                    )
                }
            )
        },
        bottomBar = {
            NavigationBar {
                val navBackStackEntry by navController.currentBackStackEntryAsState()
                val currentDestination = navBackStackEntry?.destination

                navigationItems.forEach { item ->
                    NavigationBarItem(
                        icon = { Icon(item.iconData, contentDescription = null) },
                        label = { Text(stringResource(item.labelResId)) },
                        selected = currentDestination?.hierarchy?.any { it.route == item.route } == true,
                        onClick = {
                            navController.navigate(item.route) {
                                popUpTo(navController.graph.findStartDestination().id) {
                                    saveState = true
                                }
                                launchSingleTop = true
                                restoreState = true
                            }
                        }
                    )
                }
            }
        }
    ) { innerPadding ->
        NavHost(
            navController = navController,
            startDestination = NavigationItem.Backup.route,
            modifier = Modifier.padding(innerPadding)
        ) {
            composable(NavigationItem.Backup.route) {
                BackupScreen(
                    permissionsGranted = backupViewModel.getPermissionsGranted(),
                    isOperating = backupViewModel.isOperating(),
                    backupStatus = backupViewModel.getBackupStatus(),
                    onBackupClick = { backupViewModel.startBackup() }
                )
            }
            composable(NavigationItem.Restore.route) {
                val backupFiles by restoreViewModel.backupFiles.collectAsState(initial = emptyList())
                val selectedBackupFile by restoreViewModel.selectedBackupFile.collectAsState(initial = null)

                RestoreScreen(
                    backupFiles = backupFiles,
                    isOperating = restoreViewModel.isOperating.value,
                    restoreStatus = restoreViewModel.restoreStatus.value,
                    onRestoreClick = { backupFile ->
                        restoreViewModel.restoreBackupFile(backupFile)
                    },
                    onBackupItemClick = { backupFile ->
                        restoreViewModel.selectBackupFile(backupFile)
                    },
                    selectedBackupFile = selectedBackupFile,
                    viewModel = restoreViewModel
                )
            }
            composable(NavigationItem.Preview.route) {
                PreviewScreen(permissionsGranted = backupViewModel.getPermissionsGranted())
            }
            composable(NavigationItem.More.route) {
                MoreScreen(
                    onNavigateToRestore = {
                        navController.navigate(NavigationItem.Restore.route) {
                            popUpTo(navController.graph.findStartDestination().id) {
                                saveState = true
                            }
                            launchSingleTop = true
                            restoreState = true
                        }
                    }
                )
            }
        }
    }
}
