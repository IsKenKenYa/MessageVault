package com.iskenkenya.commory.mobile.ui.screens

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Cloud
import androidx.compose.material.icons.filled.Storage
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.iskenkenya.commory.mobile.R
import com.iskenkenya.commory.mobile.runtime.RuntimeMode

@Composable
fun ModeSelectionScreen(
    onSelectMode: (RuntimeMode) -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Text(
            text = stringResource(R.string.mode_title),
            style = MaterialTheme.typography.headlineMedium,
            textAlign = TextAlign.Center
        )
        Spacer(modifier = Modifier.height(12.dp))
        Text(
            text = stringResource(R.string.mode_subtitle),
            style = MaterialTheme.typography.bodyLarge,
            textAlign = TextAlign.Center
        )
        Spacer(modifier = Modifier.height(24.dp))

        ModeCard(
            icon = { Icon(Icons.Default.Storage, contentDescription = null) },
            title = stringResource(R.string.mode_local_title),
            body = stringResource(R.string.mode_local_body),
            action = {
                Button(onClick = { onSelectMode(RuntimeMode.LOCAL_ONLY) }) {
                    Text(stringResource(R.string.mode_use_local))
                }
            }
        )
        Spacer(modifier = Modifier.height(16.dp))
        ModeCard(
            icon = { Icon(Icons.Default.Cloud, contentDescription = null) },
            title = stringResource(R.string.mode_server_title),
            body = stringResource(R.string.mode_server_body),
            action = {
                OutlinedButton(onClick = { onSelectMode(RuntimeMode.COMMORY_SERVER) }) {
                    Text(stringResource(R.string.mode_connect_server))
                }
            }
        )
    }
}

@Composable
private fun ModeCard(
    icon: @Composable () -> Unit,
    title: String,
    body: String,
    action: @Composable () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)
    ) {
        Column(modifier = Modifier.padding(18.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                icon()
                Text(
                    text = title,
                    style = MaterialTheme.typography.titleLarge,
                    modifier = Modifier.padding(start = 12.dp)
                )
            }
            Spacer(modifier = Modifier.height(8.dp))
            Text(text = body, style = MaterialTheme.typography.bodyMedium)
            Spacer(modifier = Modifier.height(14.dp))
            action()
        }
    }
}
