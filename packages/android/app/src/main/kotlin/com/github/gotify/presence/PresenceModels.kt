package com.github.gotify.presence

/** Wire values shared with the server's `/presence` API. */
enum class PresenceState(val wire: String) {
    AT_DESK("at_desk"),
    AWAY("away"),
    UNKNOWN("unknown");

    companion object {
        fun fromWire(value: String?): PresenceState = values().firstOrNull { it.wire == value } ?: UNKNOWN
    }
}

enum class PresenceSource(val wire: String) {
    MANUAL("manual"),
    LOCATION("location"),
    WIFI("wifi"),
    BLUETOOTH("bluetooth"),
    MIXED("mixed");
}

/**
 * Result of one signal. `configured` is false when the user never set this source up.
 * `available` is false when the radio is off or the permission was denied; an unavailable
 * signal must never be read as "away".
 */
data class Signal(
    val configured: Boolean = false,
    val available: Boolean = false,
    val atDesk: Boolean = false,
    val label: String? = null
)

data class PresenceSnapshot(
    val manual: Signal = Signal(),
    val location: Signal = Signal(),
    val wifi: Signal = Signal(),
    val bluetooth: Signal = Signal()
)

data class Decision(val state: PresenceState, val source: PresenceSource, val detail: String? = null)

data class Evaluation(val decision: Decision, val awayCandidateSince: Long?)

/**
 * Pure rule engine. The same vectors must be used by the iOS port.
 *
 * 1. an active manual override wins;
 * 2. any configured signal reporting at-desk wins immediately;
 * 3. all configured signals available and reporting away produce AWAY after the debounce;
 * 4. everything else is UNKNOWN.
 */
object PresenceEvaluator {
    fun evaluate(
        snapshot: PresenceSnapshot,
        awayDebounceSeconds: Int,
        now: Long,
        awayCandidateSince: Long?
    ): Evaluation {
        val manual = snapshot.manual
        if (manual.available) {
            val state = if (manual.atDesk) PresenceState.AT_DESK else PresenceState.AWAY
            return Evaluation(Decision(state, PresenceSource.MANUAL, manual.label), null)
        }

        val signals = listOf(
            PresenceSource.LOCATION to snapshot.location,
            PresenceSource.WIFI to snapshot.wifi,
            PresenceSource.BLUETOOTH to snapshot.bluetooth
        )
        val atDesk = signals.firstOrNull { (_, signal) ->
            signal.configured && signal.available && signal.atDesk
        }
        if (atDesk != null) {
            return Evaluation(
                Decision(PresenceState.AT_DESK, atDesk.first, atDesk.second.label),
                null
            )
        }

        val configured = signals.filter { (_, signal) -> signal.configured }
        val allAway = configured.isNotEmpty() && configured.all { (_, signal) ->
            signal.available && !signal.atDesk
        }
        if (!allAway) {
            return Evaluation(
                Decision(PresenceState.UNKNOWN, PresenceSource.MIXED),
                null
            )
        }

        val since = awayCandidateSince ?: now
        val debounceMs = awayDebounceSeconds.coerceAtLeast(0) * 1000L
        if (now - since < debounceMs) {
            return Evaluation(Decision(PresenceState.UNKNOWN, PresenceSource.MIXED), since)
        }
        val detail = configured.firstNotNullOfOrNull { (_, signal) -> signal.label }
        return Evaluation(Decision(PresenceState.AWAY, PresenceSource.MIXED, detail), since)
    }
}
