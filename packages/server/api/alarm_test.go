package api

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gotify/server/v3/auth"
	"github.com/gotify/server/v3/mode"
	"github.com/gotify/server/v3/model"
	"github.com/gotify/server/v3/test"
	"github.com/gotify/server/v3/test/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestAlarmSuite(t *testing.T) {
	suite.Run(t, new(AlarmSuite))
}

type AlarmSuite struct {
	suite.Suite
	db        *testdb.Database
	alarms    *AlarmAPI
	messages  *MessageAPI
	ctx       *gin.Context
	recorder  *httptest.ResponseRecorder
	notified  []*model.MessageExternal
	clock     time.Time
	clientSeq uint
}

func (s *AlarmSuite) BeforeTest(string, string) {
	mode.Set(mode.TestDev)
	s.db = testdb.NewDB(s.T())
	s.clock = testdb.Now
	s.notified = nil
	s.alarms = &AlarmAPI{DB: s.db, Notifier: s, now: func() time.Time { return s.clock }}
	s.messages = &MessageAPI{DB: s.db, Notifier: s, Alarms: s.alarms}
	timeNow = func() time.Time { return s.clock }
	s.resetContext()
}

func (s *AlarmSuite) AfterTest(string, string) {
	timeNow = time.Now
	s.db.Close()
}

func (s *AlarmSuite) Notify(userID uint, msg *model.MessageExternal) {
	s.notified = append(s.notified, msg)
}

func (s *AlarmSuite) Test_belowThreshold_doesNotOpen() {
	app := s.enabledApp(1, 7)
	s.post(app, `{"title":"disk","message":"80%","priority":7}`)

	assert.Equal(s.T(), 200, s.recorder.Code)
	assert.Nil(s.T(), s.notified[0].Incident)
	count, err := s.db.CountIncidentsByApplication(app.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(0), count)
}

func (s *AlarmSuite) Test_disabled_doesNotOpen() {
	app := s.db.User(1).NewAppWithToken(7, "app-token")
	require.NoError(s.T(), s.alarms.SaveConfig(app.ID, &model.AlarmConfig{
		Enabled:               false,
		PriorityThreshold:     8,
		RepeatIntervalSeconds: 30,
		MaxRingSeconds:        1800,
		Stages:                defaultStages(),
	}))
	s.post(app, `{"title":"disk","message":"full","priority":10}`)

	assert.Nil(s.T(), s.notified[0].Incident)
	count, err := s.db.CountIncidentsByApplication(app.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(0), count)
}

func (s *AlarmSuite) Test_open_and_secondMessageJoins() {
	app := s.enabledApp(1, 7)
	s.post(app, `{"title":"backup","message":"first","priority":8}`)
	first := s.notified[0].Incident
	require.NotNil(s.T(), first)
	assert.Equal(s.T(), model.IncidentKindOpen, first.Kind)
	assert.Equal(s.T(), model.IncidentStateRinging, first.State)
	assert.Equal(s.T(), -1, first.CurrentStage)
	assert.Equal(s.T(), "backup", first.Title)

	s.post(app, `{"title":"backup","message":"second","priority":9}`)
	second := s.notified[1].Incident
	require.NotNil(s.T(), second)
	assert.Equal(s.T(), first.ID, second.ID)
	assert.Equal(s.T(), model.IncidentKindOpen, second.Kind)
	assert.Equal(s.T(), model.IncidentStateRinging, second.State)
	assert.Equal(s.T(), "second", second.Message)

	count, err := s.db.CountIncidentsByApplication(app.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), count)
}

func (s *AlarmSuite) Test_advanceStageZero_waits_andWrongStageConflicts() {
	app := s.enabledApp(1, 7)
	s.post(app, `{"title":"backup","message":"down","priority":8}`)
	incidentID := s.notified[0].Incident.ID

	s.advanceAsClient(1, incidentID, 0)
	assert.Equal(s.T(), 200, s.recorder.Code)
	waiting := s.notified[len(s.notified)-1].Incident
	assert.Equal(s.T(), model.IncidentStateWaiting, waiting.State)
	assert.Equal(s.T(), model.IncidentKindAdvance, waiting.Kind)
	assert.Equal(s.T(), 0, waiting.CurrentStage)
	require.NotNil(s.T(), waiting.StageDeadlineAt)
	assert.True(s.T(), waiting.StageDeadlineAt.Equal(s.clock.Add(600*time.Second)))

	s.advanceAsClient(1, incidentID, 0)
	assert.Equal(s.T(), 409, s.recorder.Code)
}

func (s *AlarmSuite) Test_deadline_reopensFromTheStart() {
	app := s.enabledApp(1, 7)
	s.post(app, `{"title":"backup","message":"down","priority":8}`)
	incidentID := s.notified[0].Incident.ID
	s.advanceAsClient(1, incidentID, 0)

	s.clock = s.clock.Add(601 * time.Second)
	restarted := &AlarmAPI{DB: s.db, Notifier: s, now: func() time.Time { return s.clock }}
	require.NoError(s.T(), restarted.Tick())

	reopened := s.notified[len(s.notified)-1].Incident
	require.NotNil(s.T(), reopened)
	assert.Equal(s.T(), model.IncidentKindReopen, reopened.Kind)
	assert.Equal(s.T(), model.IncidentStateRinging, reopened.State)
	assert.Equal(s.T(), -1, reopened.CurrentStage)
	require.NotNil(s.T(), reopened.RingUntil)
	assert.True(s.T(), reopened.RingUntil.Equal(s.clock.Add(1800*time.Second)))
}

func (s *AlarmSuite) Test_lastStage_closes() {
	app := s.enabledApp(1, 7)
	s.post(app, `{"title":"backup","message":"down","priority":8}`)
	incidentID := s.notified[0].Incident.ID
	s.advanceAsClient(1, incidentID, 0)
	s.advanceAsClient(1, incidentID, 1)

	assert.Equal(s.T(), 200, s.recorder.Code)
	closed := s.notified[len(s.notified)-1].Incident
	assert.Equal(s.T(), model.IncidentStateClosed, closed.State)
	assert.Equal(s.T(), model.IncidentKindClose, closed.Kind)
}

func (s *AlarmSuite) Test_applicationToken_advanceForbidden() {
	app := s.enabledApp(1, 7)
	s.post(app, `{"title":"backup","message":"down","priority":8}`)
	incidentID := s.notified[0].Incident.ID

	s.resetContext()
	s.ctx.Request = httptest.NewRequest("POST", "/incident/1/advance", strings.NewReader(`{"stage":0}`))
	s.ctx.Request.Header.Set("Content-Type", "application/json")
	s.ctx.Params = gin.Params{{Key: "id", Value: fmt.Sprint(incidentID)}}
	auth.RegisterApplication(s.ctx, app)
	s.alarms.Advance(s.ctx)

	assert.Equal(s.T(), 403, s.recorder.Code)
}

func (s *AlarmSuite) Test_ringingPastMax_expires() {
	app := s.enabledApp(1, 7)
	s.post(app, `{"title":"backup","message":"down","priority":8}`)

	s.clock = s.clock.Add(1800 * time.Second)
	restarted := &AlarmAPI{DB: s.db, Notifier: s, now: func() time.Time { return s.clock }}
	require.NoError(s.T(), restarted.Tick())

	expired := s.notified[len(s.notified)-1].Incident
	require.NotNil(s.T(), expired)
	assert.Equal(s.T(), model.IncidentKindExpire, expired.Kind)
	assert.Equal(s.T(), model.IncidentStateExpired, expired.State)
	assert.Nil(s.T(), expired.RingUntil)
}

func (s *AlarmSuite) enabledApp(userID uint, appID uint) *model.Application {
	app := s.db.User(userID).NewAppWithToken(appID, "app-token")
	require.NoError(s.T(), s.alarms.SaveConfig(app.ID, &model.AlarmConfig{
		Enabled:               true,
		PriorityThreshold:     8,
		RepeatIntervalSeconds: 30,
		MaxRingSeconds:        1800,
		Stages:                defaultStages(),
	}))
	return app
}

func (s *AlarmSuite) post(app *model.Application, body string) {
	s.resetContext()
	s.ctx.Request = httptest.NewRequest("POST", "/message", strings.NewReader(body))
	s.ctx.Request.Header.Set("Content-Type", "application/json")
	auth.RegisterApplication(s.ctx, app)
	s.messages.CreateMessage(s.ctx)
}

func (s *AlarmSuite) advanceAsClient(userID uint, incidentID uint, stage int) {
	s.clientSeq++
	client := &model.Client{ID: s.clientSeq, Token: fmt.Sprintf("client-%d", s.clientSeq), UserID: userID}
	require.NoError(s.T(), s.db.CreateClient(client))
	s.resetContext()
	s.ctx.Request = httptest.NewRequest("POST", "/incident/advance", strings.NewReader(fmt.Sprintf(`{"stage":%d}`, stage)))
	s.ctx.Request.Header.Set("Content-Type", "application/json")
	s.ctx.Params = gin.Params{{Key: "id", Value: fmt.Sprint(incidentID)}}
	auth.RegisterClient(s.ctx, client)
	s.alarms.Advance(s.ctx)
}

func (s *AlarmSuite) resetContext() {
	s.recorder = httptest.NewRecorder()
	s.ctx, _ = gin.CreateTestContext(s.recorder)
	s.ctx.Request = httptest.NewRequest("GET", "/irrelevant", nil)
}

func (s *AlarmSuite) Test_listReturnsActiveIncident() {
	app := s.enabledApp(1, 7)
	s.post(app, `{"title":"backup","message":"down","priority":8}`)
	s.resetContext()
	test.WithUser(s.ctx, 1)
	s.alarms.List(s.ctx)
	assert.Equal(s.T(), 200, s.recorder.Code)
	assert.Contains(s.T(), s.recorder.Body.String(), `"state":"ringing"`)
}
