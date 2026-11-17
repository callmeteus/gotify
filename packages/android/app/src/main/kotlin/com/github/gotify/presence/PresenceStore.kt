package com.github.gotify.presence

import android.content.Context
import androidx.preference.PreferenceManager

/** Local configuration and last known state of the presence feature. */
internal class PresenceStore(context: Context) {
    private val prefs = PreferenceManager.getDefaultSharedPreferences(context)

    var enabled: Boolean
        get() = prefs.getBoolean(KEY_ENABLED, false)
        set(value) = prefs.edit().putBoolean(KEY_ENABLED, value).apply()

    var wifiSsids: Set<String>
        get() = splitList(prefs.getString(KEY_WIFI, ""))
        set(value) = prefs.edit().putString(KEY_WIFI, value.joinToString(",")).apply()

    var bluetoothDevices: Set<String>
        get() = splitList(prefs.getString(KEY_BLUETOOTH, ""))
        set(value) = prefs.edit().putString(KEY_BLUETOOTH, value.joinToString(",")).apply()

    /** Serialized as `latitude,longitude,radiusMeters;...`. */
    var geofences: String
        get() = prefs.getString(KEY_GEOFENCES, "") ?: ""
        set(value) = prefs.edit().putString(KEY_GEOFENCES, value).apply()

    var awayDebounceSeconds: Int
        get() = prefs.getString(KEY_DEBOUNCE, "120")?.toIntOrNull() ?: 120
        set(value) = prefs.edit().putString(KEY_DEBOUNCE, value.toString()).apply()

    var presenceTtlSeconds: Int
        get() = prefs.getString(KEY_TTL, "1800")?.toIntOrNull() ?: 1800
        set(value) = prefs.edit().putString(KEY_TTL, value.toString()).apply()

    var awayCandidateSince: Long?
        get() {
            val value = prefs.getLong(KEY_AWAY_SINCE, 0L)
            return if (value == 0L) null else value
        }
        set(value) = prefs.edit().putLong(KEY_AWAY_SINCE, value ?: 0L).apply()

    var manualState: PresenceState?
        get() = PresenceState.fromWire(prefs.getString(KEY_MANUAL_STATE, null))
        set(value) = prefs.edit().putString(KEY_MANUAL_STATE, value?.wire).apply()

    var manualUntil: Long
        get() = prefs.getLong(KEY_MANUAL_UNTIL, 0L)
        set(value) = prefs.edit().putLong(KEY_MANUAL_UNTIL, value).apply()

    var current: Decision?
        get() {
            val state = prefs.getString(KEY_CURRENT_STATE, null) ?: return null
            val source = prefs.getString(KEY_CURRENT_SOURCE, null) ?: return null
            return Decision(
                PresenceState.fromWire(state),
                PresenceSource.values().firstOrNull { it.wire == source } ?: PresenceSource.MIXED,
                prefs.getString(KEY_CURRENT_DETAIL, null)
            )
        }
        set(value) = prefs.edit().apply {
            if (value == null) {
                remove(KEY_CURRENT_STATE)
                remove(KEY_CURRENT_SOURCE)
                remove(KEY_CURRENT_DETAIL)
            } else {
                putString(KEY_CURRENT_STATE, value.state.wire)
                putString(KEY_CURRENT_SOURCE, value.source.wire)
                putString(KEY_CURRENT_DETAIL, value.detail)
            }
        }.apply()

    fun manualSignal(now: Long): Signal {
        val state = manualState ?: return Signal()
        if (manualUntil <= now) {
            manualState = null
            manualUntil = 0L
            return Signal()
        }
        return Signal(configured = true, available = true, atDesk = state == PresenceState.AT_DESK)
    }

    fun clearManual() {
        manualState = null
        manualUntil = 0L
    }

    fun parsedGeofences(): List<Geofence> = geofences.split(';').mapNotNull { part ->
        val pieces = part.split(',')
        if (pieces.size != 3) return@mapNotNull null
        val lat = pieces[0].toDoubleOrNull() ?: return@mapNotNull null
        val lon = pieces[1].toDoubleOrNull() ?: return@mapNotNull null
        val radius = pieces[2].toDoubleOrNull() ?: return@mapNotNull null
        Geofence(lat, lon, radius.toFloat())
    }

    data class Geofence(val latitude: Double, val longitude: Double, val radiusMeters: Float)

    private fun splitList(raw: String?): Set<String> = raw.orEmpty()
        .split(',')
        .map { it.trim() }
        .filter { it.isNotEmpty() }
        .toSet()

    private companion object {
        const val KEY_ENABLED = "presence_enabled"
        const val KEY_WIFI = "presence_wifi_ssids"
        const val KEY_BLUETOOTH = "presence_bluetooth_devices"
        const val KEY_GEOFENCES = "presence_geofences"
        const val KEY_DEBOUNCE = "presence_away_debounce_seconds"
        const val KEY_TTL = "presence_ttl_seconds"
        const val KEY_AWAY_SINCE = "presence_away_candidate_since"
        const val KEY_MANUAL_STATE = "presence_manual_state"
        const val KEY_MANUAL_UNTIL = "presence_manual_until"
        const val KEY_CURRENT_STATE = "presence_current_state"
        const val KEY_CURRENT_SOURCE = "presence_current_source"
        const val KEY_CURRENT_DETAIL = "presence_current_detail"
    }
}
