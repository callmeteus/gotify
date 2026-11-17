package model

import "time"

const (
	// PresenceStateAtDesk means at least one configured signal reports the user at the computer.
	PresenceStateAtDesk = "at_desk"
	// PresenceStateAway means every configured signal reports the user away after the debounce.
	PresenceStateAway = "away"
	// PresenceStateUnknown is used when no signal is available or the last report expired.
	PresenceStateUnknown = "unknown"

	// PresenceSourceManual is an explicit override from the user.
	PresenceSourceManual = "manual"
	// PresenceSourceLocation is a geofence result.
	PresenceSourceLocation = "location"
	// PresenceSourceWifi is a connected SSID result.
	PresenceSourceWifi = "wifi"
	// PresenceSourceBluetooth is a paired/connected device result.
	PresenceSourceBluetooth = "bluetooth"
	// PresenceSourceMixed means several signals contributed to the result.
	PresenceSourceMixed = "mixed"
)

// Presence is the latest presence report of one user.
type Presence struct {
	UserID     uint       `gorm:"primaryKey"`
	State      string     `gorm:"index"`
	Source     string
	Detail     string     `gorm:"type:text"`
	ReportedAt time.Time
	ValidUntil *time.Time
}

// PresenceExternal is the presence payload on the API and on the WebSocket frame.
//
// swagger:model Presence
type PresenceExternal struct {
	UserID     uint       `json:"userId"`
	State      string     `json:"state"`
	Source     string     `json:"source"`
	Detail     *string    `json:"detail,omitempty"`
	ReportedAt time.Time  `json:"reportedAt"`
	ValidUntil *time.Time `json:"validUntil,omitempty"`
}

// PresenceFrame is the opt-in frame sent on /stream?presence=1.
type PresenceFrame struct {
	Type string            `json:"type"`
	Data *PresenceExternal `json:"data"`
}
