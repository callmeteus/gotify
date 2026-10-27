package api

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gotify/server/v3/auth"
	"github.com/gotify/server/v3/model"
	"github.com/rs/zerolog/log"
)

// AlarmDatabase stores alarm configuration and incidents.
type AlarmDatabase interface {
	GetApplicationAlarm(applicationID uint) (*model.ApplicationAlarm, error)
	SaveApplicationAlarm(alarm *model.ApplicationAlarm) error
	GetActiveIncident(applicationID uint) (*model.Incident, error)
	GetIncidentByID(id uint) (*model.Incident, error)
	CreateIncident(incident *model.Incident) error
	SaveIncident(incident *model.Incident) error
	ListActiveIncidentsByUser(userID uint, state string) ([]*model.Incident, error)
	ListDueIncidents(now time.Time) ([]*model.Incident, error)
	CountIncidentsByApplication(applicationID uint) (int64, error)
}

// AlarmAPI opens incidents from messages and moves them through stages.
type AlarmAPI struct {
	DB       AlarmDatabase
	Notifier Notifier
	now      func() time.Time
}

// ErrAlarmInvalid is returned when an enabled alarm configuration is not usable.
var ErrAlarmInvalid = errors.New("alarm configuration is invalid")

// ErrIncidentConflict is returned when a stage confirmation does not match the incident.
var ErrIncidentConflict = errors.New("incident is not waiting for that stage")

type advanceBody struct {
	Stage int `json:"stage"`
}

// Now returns the clock used by deadlines.
func (a *AlarmAPI) Now() time.Time {
	if a.now == nil {
		return time.Now()
	}
	return a.now()
}

// OnMessage opens or joins an incident when the application alarm allows it.
func (a *AlarmAPI) OnMessage(app *model.Application, message *model.MessageExternal) *model.IncidentExternal {
	if a == nil || a.DB == nil || app == nil || message == nil {
		return nil
	}
	stored, err := a.DB.GetApplicationAlarm(app.ID)
	if err != nil || stored == nil || !stored.Enabled {
		return nil
	}
	config, err := configFromStored(stored)
	if err != nil {
		log.Error().Err(err).Uint("application", app.ID).Msg("alarm config is unreadable")
		return nil
	}
	priority := 0
	if message.Priority != nil {
		priority = *message.Priority
	}
	if priority < config.PriorityThreshold {
		return nil
	}

	now := a.Now()
	active, err := a.DB.GetActiveIncident(app.ID)
	if err != nil {
		log.Error().Err(err).Uint("application", app.ID).Msg("alarm lookup failed")
		return nil
	}
	if active != nil {
		active.MessageID = message.ID
		active.Title = message.Title
		active.Body = message.Message
		active.UpdatedAt = now
		if err := a.DB.SaveIncident(active); err != nil {
			log.Error().Err(err).Uint("incident", active.ID).Msg("alarm join failed")
			return nil
		}
		kind := model.IncidentKindOpen
		if active.State == model.IncidentStateWaiting {
			kind = model.IncidentKindAdvance
		}
		return toIncidentExternal(active, config, kind)
	}

	ringUntil := now.Add(time.Duration(config.MaxRingSeconds) * time.Second)
	incident := &model.Incident{
		ApplicationID: app.ID,
		UserID:        app.UserID,
		State:         model.IncidentStateRinging,
		CurrentStage:  -1,
		MessageID:     message.ID,
		Title:         message.Title,
		Body:          message.Message,
		RingUntil:     &ringUntil,
		OpenedAt:      now,
		UpdatedAt:     now,
	}
	if err := a.DB.CreateIncident(incident); err != nil {
		log.Error().Err(err).Uint("application", app.ID).Msg("alarm open failed")
		return nil
	}
	return toIncidentExternal(incident, config, model.IncidentKindOpen)
}

// Tick promotes deadlines that have passed and notifies connected clients.
func (a *AlarmAPI) Tick() error {
	due, err := a.DB.ListDueIncidents(a.Now())
	if err != nil {
		return err
	}
	for _, incident := range due {
		if err := a.promote(incident); err != nil {
			log.Error().Err(err).Uint("incident", incident.ID).Msg("alarm deadline failed")
		}
	}
	return nil
}

// Start runs the deadline loop until the returned function is called.
func (a *AlarmAPI) Start() func() {
	stop := make(chan struct{})
	go func() {
		if err := a.Tick(); err != nil {
			log.Error().Err(err).Msg("alarm tick failed")
		}
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				if err := a.Tick(); err != nil {
					log.Error().Err(err).Msg("alarm tick failed")
				}
			}
		}
	}()
	return func() {
		close(stop)
	}
}

// List returns the caller's active incidents.
// swagger:operation GET /incident incident listIncidents
//
//	---
//	produces: [application/json]
//	security: [clientTokenAuthorizationHeader: [], clientTokenHeader: [], clientTokenQuery: [], basicAuth: []]
//	parameters:
//	- name: state
//	  in: query
//	  type: string
//	  required: false
//	responses:
//	  200:
//	    description: Ok
//	    schema:
//	      type: array
//	      items:
//	        $ref: "#/definitions/Incident"
//	  403:
//	    description: Forbidden
func (a *AlarmAPI) List(ctx *gin.Context) {
	if auth.GetApplication(ctx) != nil {
		ctx.AbortWithError(403, errors.New("application tokens cannot read incidents"))
		return
	}
	state := ctx.Query("state")
	incidents, err := a.DB.ListActiveIncidentsByUser(auth.GetUserID(ctx), state)
	if success := successOrAbort(ctx, 500, err); !success {
		return
	}
	out := make([]*model.IncidentExternal, 0, len(incidents))
	for _, incident := range incidents {
		external := a.externalFor(incident, "")
		if external != nil {
			out = append(out, external)
		}
	}
	ctx.JSON(200, out)
}

// Advance confirms the next stage of an incident.
// swagger:operation POST /incident/{id}/advance incident advanceIncident
//
//	---
//	consumes: [application/json]
//	produces: [application/json]
//	security: [clientTokenAuthorizationHeader: [], clientTokenHeader: [], clientTokenQuery: [], basicAuth: []]
//	parameters:
//	- name: id
//	  in: path
//	  required: true
//	  type: integer
//	- name: body
//	  in: body
//	  required: true
//	  schema:
//	    type: object
//	    properties:
//	      stage:
//	        type: integer
//	responses:
//	  200:
//	    description: Ok
//	    schema:
//	      $ref: "#/definitions/Incident"
//	  403:
//	    description: Forbidden
//	  409:
//	    description: Conflict
func (a *AlarmAPI) Advance(ctx *gin.Context) {
	if auth.GetApplication(ctx) != nil {
		ctx.AbortWithError(403, errors.New("application tokens cannot advance incidents"))
		return
	}
	withID(ctx, "id", func(id uint) {
		body := advanceBody{}
		if err := ctx.Bind(&body); err != nil {
			return
		}
		external, err := a.advance(auth.GetUserID(ctx), id, body.Stage)
		if errors.Is(err, ErrIncidentConflict) {
			ctx.AbortWithError(409, err)
			return
		}
		if success := successOrAbort(ctx, 500, err); !success {
			return
		}
		if external == nil {
			ctx.AbortWithError(404, errors.New("incident not found"))
			return
		}
		ctx.JSON(200, external)
	})
}

// SaveConfig validates and stores an application alarm configuration.
func (a *AlarmAPI) SaveConfig(applicationID uint, config *model.AlarmConfig) error {
	if config == nil {
		return nil
	}
	if config.Enabled {
		if err := validateAlarm(config); err != nil {
			return err
		}
	}
	if config.Stages == nil {
		config.Stages = defaultStages()
	}
	raw, err := json.Marshal(config.Stages)
	if err != nil {
		return err
	}
	return a.DB.SaveApplicationAlarm(&model.ApplicationAlarm{
		ApplicationID:         applicationID,
		Enabled:               config.Enabled,
		PriorityThreshold:     config.PriorityThreshold,
		RepeatIntervalSeconds: config.RepeatIntervalSeconds,
		MaxRingSeconds:        config.MaxRingSeconds,
		Stages:                string(raw),
	})
}

// ConfigFor returns the stored alarm configuration for an application.
func (a *AlarmAPI) ConfigFor(applicationID uint) (*model.AlarmConfig, error) {
	stored, err := a.DB.GetApplicationAlarm(applicationID)
	if err != nil || stored == nil {
		return nil, err
	}
	return configFromStored(stored)
}

func (a *AlarmAPI) advance(userID uint, incidentID uint, stage int) (*model.IncidentExternal, error) {
	incident, err := a.DB.GetIncidentByID(incidentID)
	if err != nil || incident == nil || incident.UserID != userID {
		return nil, err
	}
	expected := 0
	if incident.State == model.IncidentStateWaiting {
		expected = incident.CurrentStage + 1
	} else if incident.State != model.IncidentStateRinging {
		return nil, ErrIncidentConflict
	}
	if stage != expected {
		return nil, ErrIncidentConflict
	}
	config, err := a.ConfigFor(incident.ApplicationID)
	if err != nil || config == nil {
		return nil, err
	}
	if stage < 0 || stage >= len(config.Stages) {
		return nil, ErrIncidentConflict
	}

	now := a.Now()
	kind := model.IncidentKindAdvance
	incident.CurrentStage = stage
	incident.UpdatedAt = now
	if stage == len(config.Stages)-1 {
		incident.State = model.IncidentStateClosed
		incident.StageDeadlineAt = nil
		incident.RingUntil = nil
		kind = model.IncidentKindClose
	} else {
		if config.Stages[stage].AdvanceWithinSeconds <= 0 {
			return nil, ErrIncidentConflict
		}
		deadline := now.Add(time.Duration(config.Stages[stage].AdvanceWithinSeconds) * time.Second)
		incident.State = model.IncidentStateWaiting
		incident.StageDeadlineAt = &deadline
		incident.RingUntil = nil
	}
	if err := a.DB.SaveIncident(incident); err != nil {
		return nil, err
	}
	external := toIncidentExternal(incident, config, kind)
	a.notify(incident, external)
	return external, nil
}

func (a *AlarmAPI) promote(incident *model.Incident) error {
	config, err := a.ConfigFor(incident.ApplicationID)
	if err != nil || config == nil {
		return err
	}
	now := a.Now()
	kind := ""
	if incident.State == model.IncidentStateRinging {
		incident.State = model.IncidentStateExpired
		incident.RingUntil = nil
		incident.StageDeadlineAt = nil
		kind = model.IncidentKindExpire
	} else if incident.State == model.IncidentStateWaiting {
		ringUntil := now.Add(time.Duration(config.MaxRingSeconds) * time.Second)
		incident.State = model.IncidentStateRinging
		incident.CurrentStage = -1
		incident.StageDeadlineAt = nil
		incident.RingUntil = &ringUntil
		kind = model.IncidentKindReopen
	} else {
		return nil
	}
	incident.UpdatedAt = now
	if err := a.DB.SaveIncident(incident); err != nil {
		return err
	}
	a.notify(incident, toIncidentExternal(incident, config, kind))
	return nil
}

func (a *AlarmAPI) notify(incident *model.Incident, external *model.IncidentExternal) {
	if a.Notifier == nil || external == nil {
		return
	}
	priority := 0
	message := &model.MessageExternal{
		ID:            incident.MessageID,
		ApplicationID: incident.ApplicationID,
		Message:       incident.Body,
		Title:         incident.Title,
		Priority:      &priority,
		Date:          incident.UpdatedAt,
		Incident:      external,
	}
	a.Notifier.Notify(incident.UserID, message)
}

func (a *AlarmAPI) externalFor(incident *model.Incident, kind string) *model.IncidentExternal {
	config, err := a.ConfigFor(incident.ApplicationID)
	if err != nil || config == nil {
		return nil
	}
	return toIncidentExternal(incident, config, kind)
}

func toIncidentExternal(incident *model.Incident, config *model.AlarmConfig, kind string) *model.IncidentExternal {
	return &model.IncidentExternal{
		ID:                    incident.ID,
		ApplicationID:         incident.ApplicationID,
		State:                 incident.State,
		Kind:                  kind,
		CurrentStage:          incident.CurrentStage,
		Stages:                config.Stages,
		StageDeadlineAt:       incident.StageDeadlineAt,
		RingUntil:             incident.RingUntil,
		RepeatIntervalSeconds: config.RepeatIntervalSeconds,
		Title:                 incident.Title,
		Message:               incident.Body,
	}
}

func configFromStored(stored *model.ApplicationAlarm) (*model.AlarmConfig, error) {
	stages := []model.AlarmStage{}
	if stored.Stages != "" {
		if err := json.Unmarshal([]byte(stored.Stages), &stages); err != nil {
			return nil, err
		}
	}
	return &model.AlarmConfig{
		Enabled:               stored.Enabled,
		PriorityThreshold:     stored.PriorityThreshold,
		RepeatIntervalSeconds: stored.RepeatIntervalSeconds,
		MaxRingSeconds:        stored.MaxRingSeconds,
		Stages:                stages,
	}, nil
}

func validateAlarm(config *model.AlarmConfig) error {
	if len(config.Stages) < 2 {
		return ErrAlarmInvalid
	}
	if config.RepeatIntervalSeconds <= 0 || config.MaxRingSeconds <= 0 {
		return ErrAlarmInvalid
	}
	for index, stage := range config.Stages {
		if stage.Label == "" {
			return ErrAlarmInvalid
		}
		last := index == len(config.Stages)-1
		if !last && stage.AdvanceWithinSeconds <= 0 {
			return ErrAlarmInvalid
		}
	}
	return nil
}

func defaultStages() []model.AlarmStage {
	return []model.AlarmStage{
		{Label: "I'm up", AdvanceWithinSeconds: 600},
		{Label: "I'm at my computer", AdvanceWithinSeconds: 0},
	}
}
