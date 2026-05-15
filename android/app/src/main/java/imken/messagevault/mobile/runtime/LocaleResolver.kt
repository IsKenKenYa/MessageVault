package imken.messagevault.mobile.runtime

import android.content.Context
import android.content.res.Configuration
import java.util.Locale

object LocaleResolver {
    fun resolveLocale(environment: AppEnvironment, fallback: Locale = Locale.getDefault()): Locale {
        return when (environment.locale) {
            AppLocaleOption.SYSTEM -> fallback
            AppLocaleOption.ZH_CN -> Locale.SIMPLIFIED_CHINESE
            AppLocaleOption.EN -> Locale.ENGLISH
        }
    }

    fun wrapContext(base: Context, environment: AppEnvironment): Context {
        val locale = resolveLocale(environment, base.resources.configuration.locales[0] ?: Locale.getDefault())
        Locale.setDefault(locale)
        val configuration = Configuration(base.resources.configuration)
        configuration.setLocale(locale)
        return base.createConfigurationContext(configuration)
    }
}
