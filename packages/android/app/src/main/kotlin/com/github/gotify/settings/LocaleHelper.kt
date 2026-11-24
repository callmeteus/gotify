package com.github.gotify.settings

import android.content.Context
import android.content.res.Configuration
import android.os.Build
import androidx.preference.PreferenceManager
import java.util.Locale

/** Keeps the interface in pt-BR or en-US, following the system by default. */
internal object LocaleHelper {
    private const val KEY = "language"
    private const val SYSTEM = "system"
    private const val PT = "pt-BR"
    private const val EN = "en-US"

    fun apply(context: Context) {
        val selected = PreferenceManager.getDefaultSharedPreferences(context)
            .getString(KEY, SYSTEM) ?: SYSTEM
        val locale = when (selected) {
            PT -> Locale("pt", "BR")
            EN -> Locale.US
            else -> systemLocale()
        }
        Locale.setDefault(locale)
        val config = Configuration(context.resources.configuration)
        config.setLocale(locale)
        @Suppress("DEPRECATION")
        context.resources.updateConfiguration(config, context.resources.displayMetrics)
    }

    fun systemLocale(): Locale =
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N) {
            Locale.getDefault().let { Locale(it.language, it.country) }
        } else {
            @Suppress("DEPRECATION")
            Locale.getDefault()
        }

    fun current(context: Context): String =
        PreferenceManager.getDefaultSharedPreferences(context).getString(KEY, SYSTEM) ?: SYSTEM

    fun effective(context: Context): Locale {
        val selected = current(context)
        return when (selected) {
            PT -> Locale("pt", "BR")
            EN -> Locale.US
            else -> systemLocale()
        }
    }

    fun supported(): List<Pair<String, String>> = listOf(SYSTEM to "", PT to PT, EN to EN)
}
