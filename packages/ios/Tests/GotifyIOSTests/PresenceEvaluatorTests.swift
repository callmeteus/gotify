import XCTest
@testable import GotifyIOS

final class PresenceEvaluatorTests: XCTestCase {
    func testManualOverrideWins() {
        var snapshot = PresenceSnapshot()
        snapshot.manual = Signal(configured: true, available: true, atDesk: false)
        snapshot.wifi = Signal(configured: true, available: true, atDesk: true)
        let result = PresenceEvaluator.evaluate(snapshot, awayDebounceSeconds: 120,
                                                now: Date(timeIntervalSince1970: 1000),
                                                awayCandidateSince: nil)
        XCTAssertEqual(result.decision.state, .away)
        XCTAssertEqual(result.decision.source, .manual)
    }

    func testAnyConfiguredSignalAtDeskWins() {
        var snapshot = PresenceSnapshot()
        snapshot.bluetooth = Signal(configured: true, available: true, atDesk: true, label: "headset")
        let result = PresenceEvaluator.evaluate(snapshot, awayDebounceSeconds: 120,
                                                now: Date(timeIntervalSince1970: 1000),
                                                awayCandidateSince: nil)
        XCTAssertEqual(result.decision.state, .atDesk)
        XCTAssertEqual(result.decision.source, .bluetooth)
    }

    func testAllSignalsAwayNeedDebounce() {
        var snapshot = PresenceSnapshot()
        snapshot.wifi = Signal(configured: true, available: true, atDesk: false)
        snapshot.location = Signal(configured: true, available: true, atDesk: false)
        let now = Date(timeIntervalSince1970: 1000)
        let first = PresenceEvaluator.evaluate(snapshot, awayDebounceSeconds: 120,
                                               now: now, awayCandidateSince: nil)
        XCTAssertEqual(first.decision.state, .unknown)
        let done = PresenceEvaluator.evaluate(snapshot, awayDebounceSeconds: 120,
                                              now: now.addingTimeInterval(120),
                                              awayCandidateSince: first.awayCandidateSince)
        XCTAssertEqual(done.decision.state, .away)
    }

    func testUnavailableSignalNeverCountsAsAway() {
        var snapshot = PresenceSnapshot()
        snapshot.wifi = Signal(configured: true, available: false, atDesk: false)
        let result = PresenceEvaluator.evaluate(snapshot, awayDebounceSeconds: 0,
                                                now: Date(timeIntervalSince1970: 1000),
                                                awayCandidateSince: nil)
        XCTAssertEqual(result.decision.state, .unknown)
    }
}
