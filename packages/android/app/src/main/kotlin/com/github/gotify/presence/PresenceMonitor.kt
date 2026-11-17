package com.github.gotify.presence

import android.content.Context
import android.os.Handler
import android.os.Looper
import com.github.gotify.Settings

/** Periodically evaluates the configured signals and publishes changes to the server. */
internal class PresenceMonitor(
    context: Context,
    private val settings: Settings,
    private val store: PresenceStore,
    private val onDecision: (Decision) -> Unit
) {
    private val detectors = PresenceDetectors(context, store)
    private val reporter = PresenceReporter(settings, store)
    private val handler = Handler(Looper.getMainLooper())
    private var running = false

    private val poll = object : Runnable {
        override fun run() {
            evaluate(force = false)
            if (running) {
                handler.postDelayed(this, POLL_INTERVAL_MS)
            }
        }
    }

    fun start() {
        if (running) return
        running = true
        // Publish once at startup even when the local state did not change: the server report
        // may have expired while the phone was offline.
        evaluate(force = true)
        handler.postDelayed(poll, POLL_INTERVAL_MS)
    }

    fun stop() {
        running = false
        handler.removeCallbacks(poll)
    }

    fun evaluateNow() = evaluate(force = false)

    fun setManual(state: PresenceState?, durationMillis: Long) {
        if (state == null) {
            store.clearManual()
        } else {
            store.manualState = state
            store.manualUntil = System.currentTimeMillis() + durationMillis
        }
        evaluate(force = true)
    }

    private fun evaluate(force: Boolean) {
        if (!settings.tokenExists()) return
        val now = System.currentTimeMillis()
        val snapshot = detectors.snapshot(now)
        // A manual override works even while automatic detection is turned off.
        if (!store.enabled && !snapshot.manual.available) return
        val result = PresenceEvaluator.evaluate(
            snapshot,
            store.awayDebounceSeconds,
            now,
            store.awayCandidateSince
        )
        store.awayCandidateSince = result.awayCandidateSince
        if (!force && store.current == result.decision) return
        store.current = result.decision
        onDecision(result.decision)
        reporter.publish(result.decision)
    }

    private companion object {
        const val POLL_INTERVAL_MS = 30_000L
    }
}
