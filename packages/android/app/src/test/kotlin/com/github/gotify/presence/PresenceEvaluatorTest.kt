package com.github.gotify.presence

import org.junit.Assert.assertEquals
import org.junit.Test

class PresenceEvaluatorTest {
    private val evaluator = PresenceEvaluator

    @Test
    fun manualOverrideWins() {
        val snapshot = PresenceSnapshot(
            manual = Signal(configured = true, available = true, atDesk = false),
            wifi = Signal(configured = true, available = true, atDesk = true)
        )
        val result = evaluator.evaluate(snapshot, 120, 1000L, null)
        assertEquals(PresenceState.AWAY, result.decision.state)
        assertEquals(PresenceSource.MANUAL, result.decision.source)
    }

    @Test
    fun anyConfiguredSignalAtDeskWins() {
        val snapshot = PresenceSnapshot(
            wifi = Signal(configured = true, available = true, atDesk = false),
            bluetooth = Signal(configured = true, available = true, atDesk = true, label = "headset")
        )
        val result = evaluator.evaluate(snapshot, 120, 1000L, null)
        assertEquals(PresenceState.AT_DESK, result.decision.state)
        assertEquals(PresenceSource.BLUETOOTH, result.decision.source)
        assertEquals("headset", result.decision.detail)
    }

    @Test
    fun allSignalsAwayNeedDebounce() {
        val snapshot = PresenceSnapshot(
            wifi = Signal(configured = true, available = true, atDesk = false),
            location = Signal(configured = true, available = true, atDesk = false)
        )
        val first = evaluator.evaluate(snapshot, 120, 1000L, null)
        assertEquals(PresenceState.UNKNOWN, first.decision.state)
        assertEquals(1000L, first.awayCandidateSince)

        val tooEarly = evaluator.evaluate(snapshot, 120, 1000L + 119_000L, first.awayCandidateSince)
        assertEquals(PresenceState.UNKNOWN, tooEarly.decision.state)

        val done = evaluator.evaluate(snapshot, 120, 1000L + 120_000L, first.awayCandidateSince)
        assertEquals(PresenceState.AWAY, done.decision.state)
        assertEquals(PresenceSource.MIXED, done.decision.source)
    }

    @Test
    fun unavailableSignalNeverCountsAsAway() {
        val snapshot = PresenceSnapshot(
            wifi = Signal(configured = true, available = false, atDesk = false),
            bluetooth = Signal(configured = true, available = true, atDesk = false)
        )
        val result = evaluator.evaluate(snapshot, 0, 1000L, null)
        assertEquals(PresenceState.UNKNOWN, result.decision.state)
    }

    @Test
    fun noConfiguredSignalIsUnknown() {
        val result = evaluator.evaluate(PresenceSnapshot(), 120, 1000L, null)
        assertEquals(PresenceState.UNKNOWN, result.decision.state)
    }

    @Test
    fun atDeskResetsAwayCandidate() {
        val away = PresenceSnapshot(wifi = Signal(configured = true, available = true, atDesk = false))
        val atDesk = PresenceSnapshot(wifi = Signal(configured = true, available = true, atDesk = true))
        val first = evaluator.evaluate(away, 120, 1000L, null)
        val reset = evaluator.evaluate(atDesk, 120, 1100L, first.awayCandidateSince)
        assertEquals(PresenceState.AT_DESK, reset.decision.state)
        assertEquals(null, reset.awayCandidateSince)
    }
}
