package api

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gotify/server/v3/auth"
	"github.com/gotify/server/v3/model"
)

const (
	// defaultPresenceValidFor is used when a report does not carry a validity.
	defaultPresenceValidFor = 1800 * time.Second
	// maxPresenceValidFor caps how long a single report can stay valid.
	maxPresenceValidFor = 4 * time.Hour
	// maxPresenceDetailLength keeps the free-form signal label bounded.
	maxPresenceDetailLength = 200
)

// PresenceDatabase stores the latest presence report per user.
type PresenceDatabase interface {
	GetPresence(userID uint) (*model.Presence, error)
	SavePresence(presence *model.Presence) error
	ListPresence() ([]*model.Presence, error)
}

// PresenceNotifier broadcasts a presence change to connected clients.
type PresenceNotifier interface {
	NotifyPresence(presence *model.PresenceExternal)
}

// PresenceAPI publishes and reads presence.
type PresenceAPI struct {
	DB       PresenceDatabase
	Notifier PresenceNotifier
	now      func() time.Time
}

// ErrPresenceInvalid is returned when a report has an unknown state or source.
var ErrPresenceInvalid = errors.New("presence report is invalid")

type presenceBody struct {
	State           string `json:"state"`
	Source          string `json:"source"`
	Detail          string `json:"detail"`
	ValidForSeconds int    `json:"validForSeconds"`
}

// Now returns the clock used for validity checks.
func (a *PresenceAPI) Now() time.Time {
	if a.now == nil {
		return time.Now()
	}
	return a.now()
}

// Put publishes the caller's own presence.
// swagger:operation PUT /presence presence putPresence
//
//	---
//	consumes: [application/json]
//	produces: [application/json]
//	security: [clientTokenAuthorizationHeader: [], clientTokenHeader: [], clientTokenQuery: [], basicAuth: []]
//	parameters:
//	- name: body
//	  in: body
//	  required: true
//	  schema:
//	    type: object
//	    properties:
//	      state:
//	        type: string
//	      source:
//	        type: string
//	      detail:
//	        type: string
//	      validForSeconds:
//	        type: integer
//	responses:
//	  200:
//	    description: Ok
//	    schema:
//	      $ref: "#/definitions/Presence"
//	  400:
//	    description: Bad Request
//	    schema:
//	      $ref: "#/definitions/Error"
//	  403:
//	    description: Forbidden
//	    schema:
//	      $ref: "#/definitions/Error"
func (a *PresenceAPI) Put(ctx *gin.Context) {
	if auth.GetApplication(ctx) != nil {
		ctx.AbortWithError(403, errors.New("application tokens cannot publish presence"))
		return
	}
	body := presenceBody{}
	if err := ctx.Bind(&body); err != nil {
		return
	}
	external, err := a.set(auth.GetUserID(ctx), body)
	if errors.Is(err, ErrPresenceInvalid) {
		ctx.AbortWithError(400, err)
		return
	}
	if success := successOrAbort(ctx, 500, err); !success {
		return
	}
	ctx.JSON(200, external)
}

// List returns the presence of every user.
// swagger:operation GET /presence presence listPresence
//
//	---
//	produces: [application/json]
//	security: [clientTokenAuthorizationHeader: [], clientTokenHeader: [], clientTokenQuery: [], basicAuth: []]
//	responses:
//	  200:
//	    description: Ok
//	    schema:
//	      type: array
//	      items:
//	        $ref: "#/definitions/Presence"
//	  403:
//	    description: Forbidden
//	    schema:
//	      $ref: "#/definitions/Error"
func (a *PresenceAPI) List(ctx *gin.Context) {
	if auth.GetApplication(ctx) != nil {
		ctx.AbortWithError(403, errors.New("application tokens cannot read presence"))
		return
	}
	presences, err := a.DB.ListPresence()
	if success := successOrAbort(ctx, 500, err); !success {
		return
	}
	out := make([]*model.PresenceExternal, 0, len(presences))
	now := a.Now()
	for _, presence := range presences {
		out = append(out, toPresenceExternal(presence, now))
	}
	ctx.JSON(200, out)
}

func (a *PresenceAPI) set(userID uint, body presenceBody) (*model.PresenceExternal, error) {
	if !validPresenceState(body.State) || !validPresenceSource(body.Source) || len(body.Detail) > maxPresenceDetailLength {
		return nil, ErrPresenceInvalid
	}
	validFor := time.Duration(body.ValidForSeconds) * time.Second
	if validFor <= 0 {
		validFor = defaultPresenceValidFor
	}
	if validFor > maxPresenceValidFor {
		validFor = maxPresenceValidFor
	}
	now := a.Now()
	validUntil := now.Add(validFor)
	current, err := a.DB.GetPresence(userID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		current = &model.Presence{UserID: userID}
	}
	current.State = body.State
	current.Source = body.Source
	current.Detail = body.Detail
	current.ReportedAt = now
	current.ValidUntil = &validUntil
	if err := a.DB.SavePresence(current); err != nil {
		return nil, err
	}
	external := toPresenceExternal(current, now)
	if a.Notifier != nil {
		a.Notifier.NotifyPresence(external)
	}
	return external, nil
}

func toPresenceExternal(presence *model.Presence, now time.Time) *model.PresenceExternal {
	state := presence.State
	if !validPresenceState(state) {
		state = model.PresenceStateUnknown
	}
	if presence.ValidUntil != nil && !presence.ValidUntil.After(now) {
		state = model.PresenceStateUnknown
	}
	external := &model.PresenceExternal{
		UserID:     presence.UserID,
		State:      state,
		Source:     presence.Source,
		ReportedAt: presence.ReportedAt,
		ValidUntil: presence.ValidUntil,
	}
	if presence.Detail != "" {
		detail := presence.Detail
		external.Detail = &detail
	}
	return external
}

func validPresenceState(state string) bool {
	switch state {
	case model.PresenceStateAtDesk, model.PresenceStateAway, model.PresenceStateUnknown:
		return true
	default:
		return false
	}
}

func validPresenceSource(source string) bool {
	switch source {
	case model.PresenceSourceManual, model.PresenceSourceLocation, model.PresenceSourceWifi,
		model.PresenceSourceBluetooth, model.PresenceSourceMixed:
		return true
	default:
		return false
	}
}
