package com.iskenkenya.commory.mobile.ui.screens

import android.Manifest
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import com.iskenkenya.commory.mobile.R
import com.iskenkenya.commory.mobile.ui.theme.CommoryTheme
import com.iskenkenya.commory.mobile.ui.model.UiText
import com.iskenkenya.commory.mobile.ui.model.resolve

/**
 * 备份屏幕
 *
 * 显示备份功能界面，包括备份按钮和状态信息
 *
 * @param permissionsGranted 是否已授权所需权限
 * @param isOperating 是否有操作正在进行
 * @param backupStatus 备份状态信息
 * @param onBackupClick 备份按钮点击回调
 */
@Composable
fun BackupScreen(
    permissionsGranted: Boolean,
    isOperating: Boolean,
    backupStatus: UiText?,
    onBackupClick: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        // 标题
        Text(
            text = stringResource(id = R.string.backup_tab),
            style = MaterialTheme.typography.headlineMedium,
            modifier = Modifier.padding(bottom = 24.dp)
        )
        
        if (!permissionsGranted) {
            // 权限提示
            PermissionRequiredCard()
        } else {
            // 备份信息卡片
            BackupInfoCard(isOperating, backupStatus, onBackupClick)
            
            Spacer(modifier = Modifier.height(16.dp))
            
            // 备份帮助卡片
            BackupHelpCard()
        }
    }
}

/**
 * 权限提示卡片
 *
 * 当应用缺少所需权限时显示
 */
@Composable
fun PermissionRequiredCard() {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.errorContainer
        )
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Icon(
                imageVector = Icons.Filled.Warning,
                contentDescription = null,
                tint = MaterialTheme.colorScheme.error,
                modifier = Modifier
                    .size(48.dp)
                    .padding(bottom = 8.dp)
            )
            
            Text(
                text = stringResource(id = R.string.permission_required),
                style = MaterialTheme.typography.titleLarge,
                color = MaterialTheme.colorScheme.error
            )
            
            Spacer(modifier = Modifier.height(8.dp))
            
            Text(
                text = stringResource(id = R.string.permission_prompt),
                style = MaterialTheme.typography.bodyMedium,
                textAlign = TextAlign.Center
            )
            
            Spacer(modifier = Modifier.height(16.dp))
            
            // 设置按钮
            Button(
                onClick = { /* TODO: 打开应用设置 */ },
                colors = ButtonDefaults.buttonColors(
                    containerColor = MaterialTheme.colorScheme.error
                )
            ) {
                Icon(
                    imageVector = Icons.Filled.Settings,
                    contentDescription = null,
                    modifier = Modifier.padding(end = 8.dp)
                )
                Text(text = stringResource(id = R.string.settings))
            }
        }
    }
}

/**
 * 备份信息卡片
 *
 * 显示备份状态和操作按钮
 */
@Composable
fun BackupInfoCard(
    isOperating: Boolean,
    backupStatus: UiText?,
    onBackupClick: () -> Unit
) {
    val status = backupStatus.resolve() ?: stringResource(id = R.string.status_initial)
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.primaryContainer
        )
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            // 标题
            Text(
                text = stringResource(R.string.backup_data_title),
                style = MaterialTheme.typography.titleLarge,
                color = MaterialTheme.colorScheme.onPrimaryContainer
            )
            
            Spacer(modifier = Modifier.height(16.dp))
            
            // 状态信息
            Text(
                text = status,
                style = MaterialTheme.typography.bodyLarge,
                textAlign = TextAlign.Center,
                modifier = Modifier.padding(bottom = 16.dp)
            )
            
            // 备份按钮
            Button(
                onClick = onBackupClick,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp),
                enabled = !isOperating,
                colors = ButtonDefaults.buttonColors(
                    containerColor = MaterialTheme.colorScheme.primary,
                    disabledContainerColor = MaterialTheme.colorScheme.primary.copy(alpha = 0.38f)
                )
            ) {
                if (isOperating) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(24.dp),
                        color = MaterialTheme.colorScheme.onPrimary,
                        strokeWidth = 2.dp
                    )
                } else {
                    Icon(
                        imageVector = Icons.Filled.Backup,
                        contentDescription = null,
                        modifier = Modifier.padding(end = 8.dp)
                    )
                    Text(
                        text = stringResource(id = R.string.backup_button),
                        style = MaterialTheme.typography.titleMedium
                    )
                }
            }
        }
    }
}

/**
 * 备份帮助卡片
 *
 * 显示备份相关的帮助信息
 */
@Composable
fun BackupHelpCard() {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant
        )
    ) {
        Column(
            modifier = Modifier.padding(16.dp)
        ) {
            Text(
                text = stringResource(R.string.backup_about_title),
                style = MaterialTheme.typography.titleLarge
            )
            
            Spacer(modifier = Modifier.height(8.dp))
            
            val permissions = remember {
                listOf(
                    R.string.permission_sms_backup_title to R.string.permission_sms_backup_body,
                    R.string.permission_call_backup_title to R.string.permission_call_backup_body,
                    R.string.permission_contacts_backup_title to R.string.permission_contacts_backup_body
                )
            }
            
            LazyColumn {
                items(permissions) { (permissionRes, descriptionRes) ->
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(vertical = 4.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(
                            imageVector = Icons.Filled.Check,
                            contentDescription = null,
                            tint = MaterialTheme.colorScheme.primary,
                            modifier = Modifier
                                .size(24.dp)
                                .padding(end = 8.dp)
                        )
                        
                        Column {
                            Text(
                                text = stringResource(permissionRes),
                                style = MaterialTheme.typography.titleSmall
                            )
                            
                            Text(
                                text = stringResource(descriptionRes),
                                style = MaterialTheme.typography.bodySmall
                            )
                        }
                    }
                }
                
                item {
                    Divider(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(vertical = 8.dp)
                    )
                }
                
                item {
                    Text(
                        text = "备份存储位置",
                        style = MaterialTheme.typography.titleMedium,
                        modifier = Modifier.padding(vertical = 4.dp)
                    )
                    
                    Text(
                        text = "备份文件将保存在应用私有存储空间，卸载应用会同时删除备份。请定期将备份导出到其他位置。",
                        style = MaterialTheme.typography.bodyMedium
                    )
                }
            }
        }
    }
}

/**
 * 预览函数
 */
@Preview(showBackground = true)
@Composable
fun BackupScreenPreview() {
    CommoryTheme {
        BackupScreen(
            permissionsGranted = true,
            isOperating = false,
            backupStatus = UiText.Dynamic("上次备份：2023-01-01 12:00"),
            onBackupClick = {}
        )
    }
} 
