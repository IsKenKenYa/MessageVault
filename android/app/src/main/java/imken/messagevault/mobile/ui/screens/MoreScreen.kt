package imken.messagevault.mobile.ui.screens

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Cloud
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.Language
import androidx.compose.material.icons.filled.Lock
import androidx.compose.material.icons.filled.Restore
import androidx.compose.material.icons.filled.Storage
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.Divider
import androidx.compose.material3.FilterChip
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import imken.messagevault.mobile.BuildConfig
import imken.messagevault.mobile.R
import imken.messagevault.mobile.runtime.AppEnvironment
import imken.messagevault.mobile.runtime.AppLocaleOption
import imken.messagevault.mobile.runtime.RuntimeMode
import imken.messagevault.mobile.ui.model.resolve
import imken.messagevault.mobile.ui.viewmodels.ServerAuthUiState

@Composable
fun MoreScreen(
    environment: AppEnvironment,
    serverAuthState: ServerAuthUiState,
    onModeChange: (RuntimeMode) -> Unit,
    onLocaleChange: (AppLocaleOption) -> Unit,
    onSyncOnBackupChange: (Boolean) -> Unit,
    onServerUrlChange: (String) -> Unit,
    onCheckServer: () -> Unit,
    onLogin: (String, String) -> Unit,
    onRegister: (String, String, String) -> Unit,
    onLogout: () -> Unit,
    onNavigateToRestore: () -> Unit = {}
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp)
    ) {
        Text(
            text = stringResource(id = R.string.more_tab),
            style = MaterialTheme.typography.headlineMedium,
            modifier = Modifier.padding(bottom = 16.dp)
        )

        QuickActionsCard(onNavigateToRestore)
        Spacer(modifier = Modifier.height(12.dp))
        GeneralSettingsCard(environment, onLocaleChange)
        Spacer(modifier = Modifier.height(12.dp))
        ServerSettingsCard(
            environment = environment,
            authState = serverAuthState,
            onModeChange = onModeChange,
            onSyncOnBackupChange = onSyncOnBackupChange,
            onServerUrlChange = onServerUrlChange,
            onCheckServer = onCheckServer,
            onLogin = onLogin,
            onRegister = onRegister,
            onLogout = onLogout
        )
        Spacer(modifier = Modifier.height(12.dp))
        PrivacyAndStorageCard()
        Spacer(modifier = Modifier.height(12.dp))
        AboutCard()
    }
}

@Composable
private fun QuickActionsCard(onNavigateToRestore: () -> Unit) {
    Card(modifier = Modifier.fillMaxWidth()) {
        Column(modifier = Modifier.padding(16.dp)) {
            SectionTitle(icon = Icons.Default.Restore, title = stringResource(R.string.quick_actions))
            Spacer(modifier = Modifier.height(8.dp))
            Button(onClick = onNavigateToRestore, modifier = Modifier.fillMaxWidth()) {
                Text(text = stringResource(id = R.string.restore_button))
            }
        }
    }
}

@Composable
private fun GeneralSettingsCard(
    environment: AppEnvironment,
    onLocaleChange: (AppLocaleOption) -> Unit
) {
    Card(modifier = Modifier.fillMaxWidth()) {
        Column(modifier = Modifier.padding(16.dp)) {
            SectionTitle(icon = Icons.Default.Language, title = stringResource(R.string.general_settings))
            Spacer(modifier = Modifier.height(8.dp))
            Text(text = stringResource(R.string.language_settings), style = MaterialTheme.typography.titleMedium)
            Row(modifier = Modifier.fillMaxWidth()) {
                LocaleChip(AppLocaleOption.SYSTEM, environment.locale, onLocaleChange, R.string.language_system)
                LocaleChip(AppLocaleOption.ZH_CN, environment.locale, onLocaleChange, R.string.language_zh)
                LocaleChip(AppLocaleOption.EN, environment.locale, onLocaleChange, R.string.language_en)
            }
        }
    }
}

@Composable
private fun LocaleChip(
    option: AppLocaleOption,
    selected: AppLocaleOption,
    onLocaleChange: (AppLocaleOption) -> Unit,
    labelRes: Int
) {
    FilterChip(
        selected = selected == option,
        onClick = { onLocaleChange(option) },
        label = { Text(stringResource(labelRes)) },
        modifier = Modifier.padding(end = 8.dp)
    )
}

@Composable
private fun ServerSettingsCard(
    environment: AppEnvironment,
    authState: ServerAuthUiState,
    onModeChange: (RuntimeMode) -> Unit,
    onSyncOnBackupChange: (Boolean) -> Unit,
    onServerUrlChange: (String) -> Unit,
    onCheckServer: () -> Unit,
    onLogin: (String, String) -> Unit,
    onRegister: (String, String, String) -> Unit,
    onLogout: () -> Unit
) {
    var serverUrl by remember(environment.serverUrl) { mutableStateOf(environment.serverUrl) }
    var userName by remember { mutableStateOf("") }
    var email by remember { mutableStateOf("") }
    var password by remember { mutableStateOf("") }
    var registering by remember { mutableStateOf(false) }
    val message = authState.message.resolve()

    Card(modifier = Modifier.fillMaxWidth()) {
        Column(modifier = Modifier.padding(16.dp)) {
            SectionTitle(icon = Icons.Default.Cloud, title = stringResource(R.string.server_settings))
            Spacer(modifier = Modifier.height(8.dp))
            Row {
                FilterChip(
                    selected = environment.mode == RuntimeMode.LOCAL_ONLY,
                    onClick = { onModeChange(RuntimeMode.LOCAL_ONLY) },
                    label = { Text(stringResource(R.string.mode_local_title)) },
                    modifier = Modifier.padding(end = 8.dp)
                )
                FilterChip(
                    selected = environment.mode == RuntimeMode.COMMORY_SERVER,
                    onClick = { onModeChange(RuntimeMode.COMMORY_SERVER) },
                    label = { Text(stringResource(R.string.mode_server_title)) }
                )
            }

            if (environment.mode == RuntimeMode.COMMORY_SERVER) {
                Spacer(modifier = Modifier.height(12.dp))
                OutlinedTextField(
                    value = serverUrl,
                    onValueChange = {
                        serverUrl = it
                        onServerUrlChange(it)
                    },
                    modifier = Modifier.fillMaxWidth(),
                    label = { Text(stringResource(R.string.server_url)) },
                    singleLine = true
                )
                Spacer(modifier = Modifier.height(8.dp))
                Row {
                    OutlinedButton(onClick = onCheckServer, enabled = !authState.busy) {
                        Text(stringResource(R.string.check_server))
                    }
                    TextButton(onClick = { onSyncOnBackupChange(!environment.syncOnBackup) }) {
                        Text(
                            if (environment.syncOnBackup) stringResource(R.string.sync_enabled)
                            else stringResource(R.string.sync_disabled)
                        )
                    }
                }
                Spacer(modifier = Modifier.height(8.dp))
                Row(modifier = Modifier.fillMaxWidth()) {
                    Text(
                        text = stringResource(R.string.sync_on_backup),
                        style = MaterialTheme.typography.bodyLarge,
                        modifier = Modifier.weight(1f)
                    )
                    Switch(checked = environment.syncOnBackup, onCheckedChange = onSyncOnBackupChange)
                }
                Divider(modifier = Modifier.padding(vertical = 10.dp))
                if (environment.authSession.isAuthenticated) {
                    Text(
                        text = stringResource(
                            R.string.signed_in_as,
                            environment.authSession.userName ?: environment.authSession.email ?: environment.authSession.userId.orEmpty()
                        ),
                        style = MaterialTheme.typography.bodyMedium
                    )
                    TextButton(onClick = onLogout) {
                        Text(stringResource(R.string.logout))
                    }
                } else {
                    Text(
                        text = if (registering) stringResource(R.string.create_account) else stringResource(R.string.sign_in),
                        style = MaterialTheme.typography.titleMedium
                    )
                    OutlinedTextField(
                        value = userName,
                        onValueChange = { userName = it },
                        modifier = Modifier.fillMaxWidth(),
                        label = { Text(stringResource(R.string.username)) },
                        singleLine = true
                    )
                    if (registering) {
                        OutlinedTextField(
                            value = email,
                            onValueChange = { email = it },
                            modifier = Modifier.fillMaxWidth(),
                            label = { Text(stringResource(R.string.email)) },
                            singleLine = true,
                            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Email)
                        )
                    }
                    OutlinedTextField(
                        value = password,
                        onValueChange = { password = it },
                        modifier = Modifier.fillMaxWidth(),
                        label = { Text(stringResource(R.string.password)) },
                        singleLine = true,
                        visualTransformation = PasswordVisualTransformation()
                    )
                    Row {
                        Button(
                            onClick = {
                                if (registering) onRegister(userName, email, password) else onLogin(userName, password)
                            },
                            enabled = userName.isNotBlank() && password.isNotBlank() && (!registering || email.isNotBlank())
                        ) {
                            Text(if (registering) stringResource(R.string.register) else stringResource(R.string.login))
                        }
                        TextButton(onClick = { registering = !registering }) {
                            Text(if (registering) stringResource(R.string.have_account) else stringResource(R.string.need_account))
                        }
                    }
                }
                if (!message.isNullOrBlank()) {
                    Text(text = message, style = MaterialTheme.typography.bodyMedium)
                }
            }
        }
    }
}

@Composable
private fun PrivacyAndStorageCard() {
    Card(modifier = Modifier.fillMaxWidth()) {
        Column(modifier = Modifier.padding(16.dp)) {
            SectionTitle(icon = Icons.Default.Lock, title = stringResource(R.string.privacy_permissions))
            Spacer(modifier = Modifier.height(8.dp))
            Text(text = stringResource(R.string.backup_permissions_summary), style = MaterialTheme.typography.bodyMedium)
            Spacer(modifier = Modifier.height(6.dp))
            SectionTitle(icon = Icons.Default.Storage, title = stringResource(R.string.storage_settings))
            Text(text = stringResource(R.string.storage_summary), style = MaterialTheme.typography.bodyMedium)
        }
    }
}

@Composable
private fun AboutCard() {
    Card(modifier = Modifier.fillMaxWidth()) {
        Column(modifier = Modifier.padding(16.dp)) {
            SectionTitle(icon = Icons.Default.Info, title = stringResource(R.string.about))
            Spacer(modifier = Modifier.height(8.dp))
            Text(text = stringResource(R.string.app_name), style = MaterialTheme.typography.titleMedium)
            Text(text = stringResource(R.string.version_format, BuildConfig.VERSION_NAME))
            Text(text = stringResource(R.string.app_about_body), style = MaterialTheme.typography.bodyMedium)
        }
    }
}

@Composable
private fun SectionTitle(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    title: String
) {
    Row {
        androidx.compose.material3.Icon(icon, contentDescription = null, modifier = Modifier.padding(end = 8.dp))
        Text(text = title, style = MaterialTheme.typography.titleLarge)
    }
}
