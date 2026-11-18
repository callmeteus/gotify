import Foundation

/// Port of the Android evaluator. Keep the same vectors in both test suites.
enum PresenceEvaluator {
    static func evaluate(
        _ snapshot: PresenceSnapshot,
        awayDebounceSeconds: Int,
        now: Date,
        awayCandidateSince: Date?
    ) -> Evaluation {
        if snapshot.manual.available {
            let state: PresenceState = snapshot.manual.atDesk ? .atDesk : .away
            return Evaluation(
                decision: Decision(state: state, source: .manual, detail: snapshot.manual.label),
                awayCandidateSince: nil
            )
        }

        let signals: [(PresenceSource, Signal)] = [
            (.location, snapshot.location),
            (.wifi, snapshot.wifi),
            (.bluetooth, snapshot.bluetooth),
        ]

        if let match = signals.first(where: { $0.1.configured && $0.1.available && $0.1.atDesk }) {
            return Evaluation(
                decision: Decision(state: .atDesk, source: match.0, detail: match.1.label),
                awayCandidateSince: nil
            )
        }

        let configured = signals.filter { $0.1.configured }
        let allAway = !configured.isEmpty && configured.allSatisfy { $0.1.available && !$0.1.atDesk }
        guard allAway else {
            return Evaluation(
                decision: Decision(state: .unknown, source: .mixed, detail: nil),
                awayCandidateSince: nil
            )
        }

        let since = awayCandidateSince ?? now
        guard now.timeIntervalSince(since) >= Double(max(0, awayDebounceSeconds)) else {
            return Evaluation(
                decision: Decision(state: .unknown, source: .mixed, detail: nil),
                awayCandidateSince: since
            )
        }
        let detail = configured.compactMap { $0.1.label }.first
        return Evaluation(
            decision: Decision(state: .away, source: .mixed, detail: detail),
            awayCandidateSince: since
        )
    }
}
