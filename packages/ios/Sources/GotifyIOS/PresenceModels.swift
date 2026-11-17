import Foundation

enum PresenceState: String, Decodable {
    case atDesk = "at_desk"
    case away
    case unknown
}

enum PresenceSource: String, Decodable {
    case manual
    case location
    case wifi
    case bluetooth
    case mixed
}

/// Wire model of GET/PUT /presence.
struct Presence: Decodable, Equatable {
    let userId: Int
    let state: PresenceState
    let source: String
    let detail: String?
    let reportedAt: Date
    let validUntil: Date?
}

struct PresenceFrame: Decodable {
    let type: String?
    let data: Presence?

    var isPresence: Bool { type == "presence" && data != nil }
}

struct Signal {
    var configured = false
    var available = false
    var atDesk = false
    var label: String?
}

struct PresenceSnapshot {
    var manual = Signal()
    var location = Signal()
    var wifi = Signal()
    var bluetooth = Signal()
}

struct Decision: Equatable {
    let state: PresenceState
    let source: PresenceSource
    let detail: String?
}

struct Evaluation {
    let decision: Decision
    let awayCandidateSince: Date?
}
