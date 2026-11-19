package com.github.gotify.alarm

import com.github.gotify.client.model.Incident

/** The stage index an incident is waiting for, or null when nothing can be confirmed. */
internal object IncidentStages {
    fun next(incident: Incident?): Int? {
        val state = incident?.state ?: return null
        if (state != "ringing" && state != "waiting") return null
        val next = if (state == "ringing") 0 else (incident.currentStage ?: -1) + 1
        val count = incident.stages?.size ?: return null
        return next.takeIf { it in 0 until count }
    }
}
