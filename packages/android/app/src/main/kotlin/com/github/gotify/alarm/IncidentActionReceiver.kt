package com.github.gotify.alarm

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import com.github.gotify.Settings
import com.github.gotify.api.ClientFactory
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import org.tinylog.kotlin.Logger

/** Notification action that confirms the next stage of an incident. */
internal class IncidentActionReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        val incidentId = intent.getLongExtra(EXTRA_ID, -1)
        val stage = intent.getIntExtra(EXTRA_STAGE, -1)
        if (incidentId < 0 || stage < 0) return

        val pending = goAsync()
        CoroutineScope(Dispatchers.IO).launch {
            try {
                val settings = Settings(context)
                if (settings.tokenExists()) {
                    ClientFactory.clientToken(settings)
                        .createService(IncidentApi::class.java)
                        .advance(incidentId, AdvanceStage(stage))
                        .execute()
                }
            } catch (err: Exception) {
                Logger.error(err, "incident advance from notification failed")
            } finally {
                pending.finish()
            }
        }
    }

    internal companion object {
        const val ACTION = "com.github.gotify.alarm.INCIDENT_ADVANCE"
        const val EXTRA_ID = "incident-id"
        const val EXTRA_STAGE = "incident-stage"
    }
}
