package imken.messagevault.mobile.ui.model

import androidx.annotation.StringRes
import androidx.compose.runtime.Composable
import androidx.compose.ui.res.stringResource

sealed interface UiText {
    data class Resource(@StringRes val resId: Int, val args: List<Any> = emptyList()) : UiText
    data class Dynamic(val value: String) : UiText
}

@Composable
fun UiText?.resolve(): String? {
    return when (this) {
        null -> null
        is UiText.Dynamic -> value
        is UiText.Resource -> stringResource(resId, *args.toTypedArray())
    }
}

data class PermissionSnapshot(
    val backupPermissionsGranted: Boolean = false,
    val restorePermissionsGranted: Boolean = false,
    val defaultSmsApp: Boolean = false
)
