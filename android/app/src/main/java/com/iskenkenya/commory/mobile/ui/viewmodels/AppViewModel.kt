package com.iskenkenya.commory.mobile.ui.viewmodels

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import com.iskenkenya.commory.mobile.R
import com.iskenkenya.commory.mobile.di.AppContainer
import com.iskenkenya.commory.mobile.remote.CommoryServerClient
import com.iskenkenya.commory.mobile.runtime.AppEnvironment
import com.iskenkenya.commory.mobile.runtime.AppEnvironmentManager
import com.iskenkenya.commory.mobile.runtime.AppLocaleOption
import com.iskenkenya.commory.mobile.runtime.AuthSession
import com.iskenkenya.commory.mobile.runtime.RuntimeMode
import com.iskenkenya.commory.mobile.ui.model.UiText
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

data class ServerAuthUiState(
    val busy: Boolean = false,
    val message: UiText? = null,
    val serverHealthy: Boolean = false
)

class AppViewModel(
    private val environmentManager: AppEnvironmentManager,
    private val serverClient: CommoryServerClient
) : ViewModel() {
    val environment: StateFlow<AppEnvironment> = environmentManager.environment.stateIn(
        scope = viewModelScope,
        started = SharingStarted.Eagerly,
        initialValue = environmentManager.currentSnapshot()
    )

    private val _serverAuthState = MutableStateFlow(ServerAuthUiState())
    val serverAuthState: StateFlow<ServerAuthUiState> = _serverAuthState.asStateFlow()

    fun selectMode(mode: RuntimeMode) {
        viewModelScope.launch {
            environmentManager.selectMode(mode)
            if (mode == RuntimeMode.LOCAL_ONLY) {
                environmentManager.clearSession()
            }
        }
    }

    fun updateServerUrl(serverUrl: String) {
        viewModelScope.launch {
            environmentManager.updateServerUrl(serverUrl)
            _serverAuthState.value = _serverAuthState.value.copy(serverHealthy = false, message = null)
        }
    }

    fun updateLocale(locale: AppLocaleOption) {
        viewModelScope.launch {
            environmentManager.updateLocale(locale)
        }
    }

    fun updateSyncOnBackup(enabled: Boolean) {
        viewModelScope.launch {
            environmentManager.updateSyncOnBackup(enabled)
        }
    }

    fun checkServer(baseUrl: String = environment.value.serverUrl) {
        viewModelScope.launch {
            _serverAuthState.value = ServerAuthUiState(busy = true)
            val result = serverClient.checkSetup(baseUrl)
            _serverAuthState.value = if (result.isSuccess) {
                ServerAuthUiState(
                    serverHealthy = true,
                    message = UiText.Resource(R.string.server_check_success)
                )
            } else {
                ServerAuthUiState(
                    serverHealthy = false,
                    message = UiText.Resource(R.string.server_check_failed, listOf(result.exceptionOrNull()?.message.orEmpty()))
                )
            }
        }
    }

    fun login(userName: String, password: String) {
        viewModelScope.launch {
            _serverAuthState.value = _serverAuthState.value.copy(busy = true, message = null)
            val env = environment.value
            val result = serverClient.login(env.serverUrl, userName, password)
            persistAuthResult(result)
        }
    }

    fun register(userName: String, email: String, password: String) {
        viewModelScope.launch {
            _serverAuthState.value = _serverAuthState.value.copy(busy = true, message = null)
            val env = environment.value
            val result = serverClient.register(env.serverUrl, userName, email, password)
            persistAuthResult(result)
        }
    }

    fun logout() {
        viewModelScope.launch {
            serverClient.logoutPersistedSession().getOrElse {
                environmentManager.clearSession()
            }
            _serverAuthState.value = ServerAuthUiState(message = UiText.Resource(R.string.logged_out))
        }
    }

    private suspend fun persistAuthResult(result: Result<AuthSession>) {
        if (result.isSuccess) {
            val session = result.getOrThrow()
            environmentManager.updateSession(session)
            _serverAuthState.value = ServerAuthUiState(
                serverHealthy = true,
                message = UiText.Resource(R.string.auth_success)
            )
        } else {
            _serverAuthState.value = ServerAuthUiState(
                busy = false,
                message = UiText.Resource(R.string.auth_failed, listOf(result.exceptionOrNull()?.message.orEmpty()))
            )
        }
    }

    class Factory(private val container: AppContainer) : ViewModelProvider.Factory {
        @Suppress("UNCHECKED_CAST")
        override fun <T : ViewModel> create(modelClass: Class<T>): T {
            if (modelClass.isAssignableFrom(AppViewModel::class.java)) {
                return AppViewModel(container.environmentManager, container.serverClient) as T
            }
            throw IllegalArgumentException("Unknown ViewModel class")
        }
    }
}
