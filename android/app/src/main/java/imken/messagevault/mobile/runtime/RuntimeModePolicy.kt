package imken.messagevault.mobile.runtime

object RuntimeModePolicy {
    fun requiresServerAuth(environment: AppEnvironment): Boolean {
        return environment.mode == RuntimeMode.COMMORY_SERVER && !environment.authSession.isAuthenticated
    }

    fun canUploadBackup(environment: AppEnvironment): Boolean {
        return environment.mode == RuntimeMode.COMMORY_SERVER &&
            environment.syncOnBackup &&
            environment.authSession.isAuthenticated
    }

    fun allowsLocalBackup(environment: AppEnvironment): Boolean = environment.modeSelected

    fun preservesLocalFilesOnModeSwitch(): Boolean = true
}
