package api

import (
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

func TestPresenceSuite(t *testing.T) {
	suite.Run(t, new(PresenceSuite))
}

type PresenceSuite struct {
	suite.Suite
	db       *testdb.Database
	presence *PresenceAPI
	ctx      *gin.Context
	recorder *httptest.ResponseRecorder
	clock    time.Time
	notified []*model.PresenceExternal
}

func (s *PresenceSuite) BeforeTest(string, string) {
	mode.Set(mode.TestDev)
	s.db = testdb.NewDB(s.T())
	s.clock = testdb.Now
	s.notified = nil
	s.presence = &PresenceAPI{DB: s.db, Notifier: s, now: func() time.Time { return s.clock }}
	s.resetContext()
}

func (s *PresenceSuite) AfterTest(string, string) {
	s.db.Close()
}

func (s *PresenceSuite) NotifyPresence(presence *model.PresenceExternal) {
	s.notified = append(s.notified, presence)
}

func (s *PresenceSuite) Test_put_storesAndNotifies() {
	s.putAsClient(1, `{"state":"at_desk","source":"wifi","detail":"office","validForSeconds":600}`)

	assert.Equal(s.T(), 200, s.recorder.Code)
	require.Len(s.T(), s.notified, 1)
	assert.Equal(s.T(), uint(1), s.notified[0].UserID)
	assert.Equal(s.T(), model.PresenceStateAtDesk, s.notified[0].State)
	assert.Equal(s.T(), model.PresenceSourceWifi, s.notified[0].Source)
	require.NotNil(s.T(), s.notified[0].Detail)
	assert.Equal(s.T(), "office", *s.notified[0].Detail)
	assert.True(s.T(), s.notified[0].ValidUntil.Equal(s.clock.Add(600*time.Second)))

	stored, err := s.db.GetPresence(1)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), stored)
	assert.Equal(s.T(), model.PresenceStateAtDesk, stored.State)
}

func (s *PresenceSuite) Test_put_replacesPreviousReport() {
	s.putAsClient(1, `{"state":"at_desk","source":"wifi","validForSeconds":600}`)
	s.putAsClient(1, `{"state":"away","source":"location","validForSeconds":600}`)

	assert.Equal(s.T(), 200, s.recorder.Code)
	stored, err := s.db.GetPresence(1)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), model.PresenceStateAway, stored.State)
	assert.Equal(s.T(), model.PresenceSourceLocation, stored.Source)
}

func (s *PresenceSuite) Test_applicationToken_isForbidden() {
	app := s.db.User(1).NewAppWithToken(7, "app-token")
	s.resetContext()
	s.ctx.Request = httptest.NewRequest("PUT", "/presence", strings.NewReader(`{"state":"at_desk","source":"wifi"}`))
	s.ctx.Request.Header.Set("Content-Type", "application/json")
	auth.RegisterApplication(s.ctx, app)
	s.presence.Put(s.ctx)

	assert.Equal(s.T(), 403, s.recorder.Code)
	assert.Empty(s.T(), s.notified)
}

func (s *PresenceSuite) Test_invalidStateAndSource_areRejected() {
	s.putAsClient(1, `{"state":"busy","source":"wifi"}`)
	assert.Equal(s.T(), 400, s.recorder.Code)

	s.putAsClient(1, `{"state":"at_desk","source":"telepathy"}`)
	assert.Equal(s.T(), 400, s.recorder.Code)

	assert.Empty(s.T(), s.notified)
}

func (s *PresenceSuite) Test_validForIsCapped() {
	s.putAsClient(1, `{"state":"at_desk","source":"manual","validForSeconds":99999}`)

	assert.Equal(s.T(), 200, s.recorder.Code)
	require.Len(s.T(), s.notified, 1)
	assert.True(s.T(), s.notified[0].ValidUntil.Equal(s.clock.Add(maxPresenceValidFor)))
}

func (s *PresenceSuite) Test_expiredReport_readsAsUnknown() {
	s.putAsClient(1, `{"state":"at_desk","source":"wifi","validForSeconds":60}`)

	s.clock = s.clock.Add(61 * time.Second)
	s.listAsClient(2)

	assert.Equal(s.T(), 200, s.recorder.Code)
	assert.Contains(s.T(), s.recorder.Body.String(), `"state":"unknown"`)
	assert.NotContains(s.T(), s.recorder.Body.String(), `"state":"at_desk"`)
}

func (s *PresenceSuite) Test_list_showsOtherUsers() {
	s.putAsClient(1, `{"state":"at_desk","source":"bluetooth","validForSeconds":600}`)
	s.listAsClient(2)

	assert.Equal(s.T(), 200, s.recorder.Code)
	assert.Contains(s.T(), s.recorder.Body.String(), `"userId":1`)
	assert.Contains(s.T(), s.recorder.Body.String(), `"state":"at_desk"`)
}

func (s *PresenceSuite) Test_list_applicationTokenIsForbidden() {
	app := s.db.User(1).NewAppWithToken(7, "app-token")
	s.resetContext()
	auth.RegisterApplication(s.ctx, app)
	s.presence.List(s.ctx)

	assert.Equal(s.T(), 403, s.recorder.Code)
}

func (s *PresenceSuite) Test_defaultValidForIsThirtyMinutes() {
	s.putAsClient(1, `{"state":"at_desk","source":"location"}`)

	assert.Equal(s.T(), 200, s.recorder.Code)
	require.Len(s.T(), s.notified, 1)
	assert.True(s.T(), s.notified[0].ValidUntil.Equal(s.clock.Add(30*time.Minute)))
}

func (s *PresenceSuite) Test_detailTooLong_isRejected() {
	body := `{"state":"at_desk","source":"wifi","detail":"` + strings.Repeat("x", maxPresenceDetailLength+1) + `"}`
	s.putAsClient(1, body)

	assert.Equal(s.T(), 400, s.recorder.Code)
	assert.Empty(s.T(), s.notified)
}

func (s *PresenceSuite) Test_unknownState_isAccepted() {
	s.putAsClient(1, `{"state":"unknown","source":"mixed"}`)

	assert.Equal(s.T(), 200, s.recorder.Code)
	require.Len(s.T(), s.notified, 1)
	assert.Equal(s.T(), model.PresenceStateUnknown, s.notified[0].State)
}

func (s *PresenceSuite) Test_reportBeforeExpiry_keepsState() {
	s.putAsClient(1, `{"state":"away","source":"bluetooth","validForSeconds":600}`)
	s.clock = s.clock.Add(599 * time.Second)
	s.listAsClient(2)

	assert.Equal(s.T(), 200, s.recorder.Code)
	assert.Contains(s.T(), s.recorder.Body.String(), `"state":"away"`)
}

func (s *PresenceSuite) Test_listWithoutReports_isEmptyArray() {
	s.listAsClient(1)

	assert.Equal(s.T(), 200, s.recorder.Code)
	assert.JSONEq(s.T(), `[]`, s.recorder.Body.String())
}

func (s *PresenceSuite) putAsClient(userID uint, body string) {
	s.resetContext()
	s.ctx.Request = httptest.NewRequest("PUT", "/presence", strings.NewReader(body))
	s.ctx.Request.Header.Set("Content-Type", "application/json")
	test.WithUser(s.ctx, userID)
	s.presence.Put(s.ctx)
}

func (s *PresenceSuite) listAsClient(userID uint) {
	s.resetContext()
	s.ctx.Request = httptest.NewRequest("GET", "/presence", nil)
	test.WithUser(s.ctx, userID)
	s.presence.List(s.ctx)
}

func (s *PresenceSuite) resetContext() {
	s.recorder = httptest.NewRecorder()
	s.ctx, _ = gin.CreateTestContext(s.recorder)
	s.ctx.Request = httptest.NewRequest("GET", "/irrelevant", nil)
}
