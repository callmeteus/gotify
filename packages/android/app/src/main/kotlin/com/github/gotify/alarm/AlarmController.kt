package com.github.gotify.alarm

import android.app.AlarmManager
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.media.AudioAttributes
import android.media.MediaPlayer
import android.media.RingtoneManager
import android.net.Uri
import android.os.Build
import android.provider.Settings
import com.github.gotify.Utils
import com.github.gotify.client.model.Message
import org.tinylog.kotlin.Logger

internal object AlarmController {
    const val ACTION_REPEAT = "com.github.gotify.alarm.REPEAT"

    private const val REPEAT_REQUEST = 4401

    var visible: Boolean = false

    private var player: MediaPlayer? = null
    private var ringing = false

    fun canUseFullScreen(context: Context): Boolean {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.UPSIDE_DOWN_CAKE) {
            return true
        }
        val manager = context.getSystemService(NotificationManager::class.java)
        return manager.canUseFullScreenIntent()
    }

    fun permissionIntent(context: Context): Intent? {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.UPSIDE_DOWN_CAKE) {
            return null
        }
        return Intent(Settings.ACTION_MANAGE_APP_USE_FULL_SCREEN_INTENT).apply {
            data = Uri.parse("package:${context.packageName}")
            flags = Intent.FLAG_ACTIVITY_NEW_TASK
        }
    }

    /**
     * Applies an incident frame.
     *
     * Returns true when the caller should also post the high-priority notification.
     */
    fun onIncident(context: Context, messageJson: String, appName: String): Boolean {
        val message = Utils.JSON.fromJson(messageJson, Message::class.java)
        val incident = message.incident ?: return false
        val kind = incident.kind ?: ""
        when (kind) {
            "expire" -> {
                stop(context)
                return false
            }
            "close", "advance" -> {
                stop(context)
                present(context, messageJson, appName, force = visible)
                return false
            }
            "reopen" -> {
                if (!canUseFullScreen(context)) {
                    return true
                }
                startRinging(context, message, messageJson, appName)
                return true
            }
            "open" -> {
                if (!canUseFullScreen(context)) {
                    return true
                }
                if (ringing) {
                    present(context, messageJson, appName, force = visible)
                    return false
                }
                startRinging(context, message, messageJson, appName)
                return true
            }
            else -> {
                if (incident.state == "ringing") {
                    if (!canUseFullScreen(context)) {
                        return true
                    }
                    startRinging(context, message, messageJson, appName)
                    return true
                }
                if (incident.state == "waiting") {
                    stop(context)
                    present(context, messageJson, appName, force = true)
                }
                return false
            }
        }
    }

    fun onRepeat(context: Context) {
        if (!ringing) {
            return
        }
        play(context)
    }

    fun stop(context: Context) {
        ringing = false
        player?.run {
            try {
                stop()
            } catch (err: IllegalStateException) {
                Logger.error(err, "alarm player stop failed")
            }
            release()
        }
        player = null
        val manager = context.getSystemService(AlarmManager::class.java)
        manager.cancel(repeatIntent(context))
    }

    private fun startRinging(
        context: Context,
        message: Message,
        messageJson: String,
        appName: String
    ) {
        ringing = true
        play(context)
        val seconds = message.incident?.repeatIntervalSeconds ?: 30
        schedule(context, seconds)
        present(context, messageJson, appName, force = canUseFullScreen(context))
    }

    private fun present(context: Context, messageJson: String, appName: String, force: Boolean) {
        if (!force && !visible) {
            return
        }
        val intent = Intent(context, AlarmActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_SINGLE_TOP
            putExtra(AlarmActivity.EXTRA_MESSAGE, messageJson)
            putExtra(AlarmActivity.EXTRA_APP_NAME, appName)
        }
        context.startActivity(intent)
    }

    private fun play(context: Context) {
        player?.release()
        player = null
        val created = MediaPlayer()
        try {
            created.setAudioAttributes(
                AudioAttributes.Builder()
                    .setUsage(AudioAttributes.USAGE_ALARM)
                    .setContentType(AudioAttributes.CONTENT_TYPE_SONIFICATION)
                    .build()
            )
            val uri = RingtoneManager.getDefaultUri(RingtoneManager.TYPE_ALARM)
                ?: Settings.System.DEFAULT_NOTIFICATION_URI
            created.setDataSource(context, uri)
            created.setOnCompletionListener {
                it.release()
                if (player === it) {
                    player = null
                }
            }
            created.prepare()
            created.start()
            player = created
        } catch (err: Exception) {
            Logger.error(err, "alarm sound failed")
            created.release()
        }
    }

    private fun schedule(context: Context, intervalSeconds: Int) {
        val manager = context.getSystemService(AlarmManager::class.java)
        val trigger = System.currentTimeMillis() + intervalSeconds.coerceAtLeast(1) * 1000L
        val pending = repeatIntent(context)
        try {
            manager.setExactAndAllowWhileIdle(AlarmManager.RTC_WAKEUP, trigger, pending)
        } catch (err: SecurityException) {
            Logger.error(err, "exact alarm was not allowed")
            manager.setAndAllowWhileIdle(AlarmManager.RTC_WAKEUP, trigger, pending)
        }
    }

    private fun repeatIntent(context: Context): PendingIntent {
        val intent = Intent(context, AlarmRepeatReceiver::class.java)
        return PendingIntent.getBroadcast(
            context,
            REPEAT_REQUEST,
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
    }
}
