package api

import (
	"strings"

	"github.com/gotify/server/v3/model"
)

// resolutionAlertName returns the alert name when a message is a Grafana
// resolution. It accepts the explicit alert::status extra as well as the ✅
// prefix used by the provisioning template in gotify/website.
func resolutionAlertName(message *model.MessageExternal) (string, bool) {
	if message == nil {
		return "", false
	}
	status := ""
	name := ""
	if message.Extras != nil {
		if raw, ok := message.Extras["alert::status"]; ok {
			status, _ = raw.(string)
		}
		if raw, ok := message.Extras["alert::name"]; ok {
			name, _ = raw.(string)
		}
	}
	if strings.EqualFold(status, "resolved") {
		return strings.TrimSpace(name), true
	}
	title := strings.TrimSpace(message.Title)
	if strings.HasPrefix(title, "✅") {
		if name == "" {
			name = strings.TrimSpace(strings.TrimPrefix(title, "✅"))
		}
		return name, true
	}
	return "", false
}

// closeOnResolution closes the active incident of an application when it
// belongs to the resolved alert. It returns the close frame to attach to the
// message, or nil when there is nothing to close.
func (a *AlarmAPI) closeOnResolution(app *model.Application, message *model.MessageExternal, alertName string) *model.IncidentExternal {
	if a == nil || a.DB == nil || app == nil {
		return nil
	}
	incident, err := a.DB.GetActiveIncident(app.ID)
	if err != nil || incident == nil {
		return nil
	}
	if alertName != "" && !strings.Contains(incident.Title, alertName) {
		return nil
	}
	config, err := a.ConfigFor(incident.ApplicationID)
	if err != nil || config == nil {
		return nil
	}
	now := a.Now()
	incident.State = model.IncidentStateClosed
	incident.StageDeadlineAt = nil
	incident.RingUntil = nil
	incident.UpdatedAt = now
	if err := a.DB.SaveIncident(incident); err != nil {
		return nil
	}
	external := toIncidentExternal(incident, config, model.IncidentKindClose)
	a.notify(incident, external)
	return external
}
