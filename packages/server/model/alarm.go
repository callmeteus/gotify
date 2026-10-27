package model

import "time"

const (
	// IncidentStateRinging is the state where the phone should ring.
	IncidentStateRinging = "ringing"
	// IncidentStateWaiting is the state where a stage was confirmed and a deadline is running.
	IncidentStateWaiting = "waiting"
	// IncidentStateClosed is the state where the last stage was confirmed.
	IncidentStateClosed = "closed"
	// IncidentStateExpired is the state where ringing ran out without any stage.
	IncidentStateExpired = "expired"

	// IncidentKindOpen is the first message that opened the incident.
	IncidentKindOpen = "open"
	// IncidentKindReopen is a deadline that sent the incident back to ringing.
	IncidentKindReopen = "reopen"
	// IncidentKindAdvance is a confirmed stage that is not the last one.
	IncidentKindAdvance = "advance"
	// IncidentKindClose is the last stage.
	IncidentKindClose = "close"
	// IncidentKindExpire is ringing that reached maxRing without a stage.
	IncidentKindExpire = "expire"
)

// AlarmStage is one button on the alarm screen.
type AlarmStage struct {
	Label                string `json:"label"`
	AdvanceWithinSeconds int    `json:"advanceWithinSeconds"`
}

// AlarmConfig is the alarm settings returned with an application.
type AlarmConfig struct {
	Enabled               bool         `json:"enabled"`
	PriorityThreshold     int          `json:"priorityThreshold"`
	RepeatIntervalSeconds int          `json:"repeatIntervalSeconds"`
	MaxRingSeconds        int          `json:"maxRingSeconds"`
	Stages                []AlarmStage `json:"stages"`
}

// ApplicationAlarm is the stored alarm configuration for one application.
type ApplicationAlarm struct {
	ApplicationID         uint `gorm:"primaryKey"`
	Enabled               bool
	PriorityThreshold     int
	RepeatIntervalSeconds int
	MaxRingSeconds        int
	Stages                string `gorm:"type:text"`
}

// Incident is one alarm lifecycle for an application.
type Incident struct {
	ID              uint `gorm:"primaryKey;autoIncrement"`
	ApplicationID   uint `gorm:"index"`
	UserID          uint `gorm:"index"`
	State           string
	CurrentStage    int
	MessageID       uint
	Title           string `gorm:"type:text"`
	Body            string `gorm:"type:text"`
	StageDeadlineAt *time.Time
	RingUntil       *time.Time
	OpenedAt        time.Time
	UpdatedAt       time.Time
}

// IncidentExternal is the alarm payload on a message and on the incident routes.
//
// swagger:model Incident
type IncidentExternal struct {
	ID                    uint         `json:"id"`
	ApplicationID         uint         `json:"appid"`
	State                 string       `json:"state"`
	Kind                  string       `json:"kind"`
	CurrentStage          int          `json:"currentStage"`
	Stages                []AlarmStage `json:"stages"`
	StageDeadlineAt       *time.Time   `json:"stageDeadlineAt,omitempty"`
	RingUntil             *time.Time   `json:"ringUntil,omitempty"`
	RepeatIntervalSeconds int          `json:"repeatIntervalSeconds"`
	Title                 string       `json:"title"`
	Message               string       `json:"message"`
}
