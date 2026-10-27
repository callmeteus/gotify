package com.github.gotify.alarm

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import androidx.core.content.ContextCompat
import com.github.gotify.service.WebSocketService

internal class AlarmRepeatReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        val service = Intent(context, WebSocketService::class.java).apply {
            action = AlarmController.ACTION_REPEAT
        }
        ContextCompat.startForegroundService(context, service)
    }
}
