package com.github.gotify.presence

import android.Manifest
import android.bluetooth.BluetoothManager
import android.bluetooth.BluetoothProfile
import android.content.Context
import android.content.pm.PackageManager
import android.location.LocationManager
import android.net.wifi.WifiManager
import androidx.core.content.ContextCompat
import kotlin.math.atan2
import kotlin.math.cos
import kotlin.math.sin
import kotlin.math.sqrt

/** Reads the configured signals from the device. Every failure becomes an unavailable signal. */
internal class PresenceDetectors(private val context: Context, private val store: PresenceStore) {

    fun snapshot(now: Long): PresenceSnapshot = PresenceSnapshot(
        manual = store.manualSignal(now),
        location = locationSignal(),
        wifi = wifiSignal(),
        bluetooth = bluetoothSignal()
    )

    private fun wifiSignal(): Signal {
        val configured = store.wifiSsids
        if (configured.isEmpty()) return Signal()
        if (!hasPermission(Manifest.permission.ACCESS_FINE_LOCATION)) return Signal(configured = true)

        return try {
            val manager = context.applicationContext.getSystemService(Context.WIFI_SERVICE) as? WifiManager
            if (manager == null || !manager.isWifiEnabled) {
                Signal(configured = true)
            } else {
                @Suppress("DEPRECATION")
                val raw = manager.connectionInfo?.ssid
                val ssid = raw?.trim('"')?.takeIf { it.isNotBlank() && it != "<unknown ssid>" }
                if (ssid == null) {
                    Signal(configured = true)
                } else {
                    Signal(configured = true, available = true, atDesk = configured.any { it.equals(ssid, ignoreCase = true) }, label = ssid)
                }
            }
        } catch (err: SecurityException) {
            Signal(configured = true)
        }
    }

    private fun bluetoothSignal(): Signal {
        val configured = store.bluetoothDevices
        if (configured.isEmpty()) return Signal()
        if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.S &&
            !hasPermission(Manifest.permission.BLUETOOTH_CONNECT)
        ) {
            return Signal(configured = true)
        }

        return try {
            val manager = context.applicationContext.getSystemService(Context.BLUETOOTH_SERVICE) as? BluetoothManager
            val adapter = manager?.adapter
            if (adapter == null || !adapter.isEnabled) {
                Signal(configured = true)
            } else {
                val connected = manager.getConnectedDevices(BluetoothProfile.GATT).orEmpty()
                val name = connected.firstNotNullOfOrNull { device ->
                    val deviceName = try {
                        device.name
                    } catch (err: SecurityException) {
                        null
                    }
                    deviceName?.takeIf { candidate -> configured.any { it.equals(candidate, ignoreCase = true) } }
                }
                Signal(
                    configured = true,
                    available = true,
                    atDesk = name != null,
                    label = name
                )
            }
        } catch (err: SecurityException) {
            Signal(configured = true)
        }
    }

    private fun locationSignal(): Signal {
        val fences = store.parsedGeofences()
        if (fences.isEmpty()) return Signal()
        if (!hasPermission(Manifest.permission.ACCESS_COARSE_LOCATION)) return Signal(configured = true)

        return try {
            val manager = context.applicationContext.getSystemService(Context.LOCATION_SERVICE) as? LocationManager
            val location = manager?.getLastKnownLocation(LocationManager.NETWORK_PROVIDER)
                ?: manager?.getLastKnownLocation(LocationManager.GPS_PROVIDER)
            if (location == null) {
                Signal(configured = true)
            } else {
                val inside = fences.firstOrNull { fence ->
                    distanceMeters(
                        location.latitude,
                        location.longitude,
                        fence.latitude,
                        fence.longitude
                    ) <= fence.radiusMeters
                }
                Signal(
                    configured = true,
                    available = true,
                    atDesk = inside != null,
                    label = if (inside != null) "geofence" else null
                )
            }
        } catch (err: SecurityException) {
            Signal(configured = true)
        }
    }

    private fun hasPermission(permission: String): Boolean =
        ContextCompat.checkSelfPermission(context, permission) == PackageManager.PERMISSION_GRANTED

    private fun distanceMeters(lat1: Double, lon1: Double, lat2: Double, lon2: Double): Float {
        val earthRadius = 6371000.0
        val dLat = Math.toRadians(lat2 - lat1)
        val dLon = Math.toRadians(lon2 - lon1)
        val a = sin(dLat / 2) * sin(dLat / 2) +
            cos(Math.toRadians(lat1)) * cos(Math.toRadians(lat2)) * sin(dLon / 2) * sin(dLon / 2)
        return (earthRadius * 2 * atan2(sqrt(a), sqrt(1 - a))).toFloat()
    }
}
