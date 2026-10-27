package database

import (
	"errors"
	"time"

	"github.com/gotify/server/v3/model"
	"gorm.io/gorm"
)

// GetApplicationAlarm returns the stored alarm configuration, or nil when it was never saved.
func (d *GormDatabase) GetApplicationAlarm(applicationID uint) (*model.ApplicationAlarm, error) {
	alarm := new(model.ApplicationAlarm)
	err := d.DB.Where("application_id = ?", applicationID).First(alarm).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return alarm, nil
}

// SaveApplicationAlarm inserts or replaces the alarm configuration for an application.
func (d *GormDatabase) SaveApplicationAlarm(alarm *model.ApplicationAlarm) error {
	return d.DB.Save(alarm).Error
}

// GetActiveIncident returns the ringing or waiting incident for an application, or nil.
func (d *GormDatabase) GetActiveIncident(applicationID uint) (*model.Incident, error) {
	incident := new(model.Incident)
	err := d.DB.Where(
		"application_id = ? AND state IN ?",
		applicationID,
		[]string{model.IncidentStateRinging, model.IncidentStateWaiting},
	).Order("id desc").First(incident).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return incident, nil
}

// GetIncidentByID returns one incident, or nil when it does not exist.
func (d *GormDatabase) GetIncidentByID(id uint) (*model.Incident, error) {
	incident := new(model.Incident)
	err := d.DB.Where("id = ?", id).First(incident).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return incident, nil
}

// CreateIncident inserts an incident and fills its id.
func (d *GormDatabase) CreateIncident(incident *model.Incident) error {
	return d.DB.Create(incident).Error
}

// SaveIncident writes the current incident row.
func (d *GormDatabase) SaveIncident(incident *model.Incident) error {
	return d.DB.Save(incident).Error
}

// ListActiveIncidentsByUser returns ringing and waiting incidents for a user.
func (d *GormDatabase) ListActiveIncidentsByUser(userID uint, state string) ([]*model.Incident, error) {
	query := d.DB.Where("user_id = ?", userID)
	if state == model.IncidentStateRinging || state == model.IncidentStateWaiting {
		query = query.Where("state = ?", state)
	} else {
		query = query.Where("state IN ?", []string{model.IncidentStateRinging, model.IncidentStateWaiting})
	}
	var incidents []*model.Incident
	err := query.Order("id desc").Find(&incidents).Error
	return incidents, err
}

// ListDueIncidents returns incidents whose ring or stage deadline has passed.
func (d *GormDatabase) ListDueIncidents(now time.Time) ([]*model.Incident, error) {
	var incidents []*model.Incident
	err := d.DB.Where(
		"(state = ? AND ring_until IS NOT NULL AND ring_until <= ?) OR (state = ? AND stage_deadline_at IS NOT NULL AND stage_deadline_at <= ?)",
		model.IncidentStateRinging,
		now,
		model.IncidentStateWaiting,
		now,
	).Find(&incidents).Error
	return incidents, err
}

// CountIncidentsByApplication returns how many incidents exist for an application.
func (d *GormDatabase) CountIncidentsByApplication(applicationID uint) (int64, error) {
	var count int64
	err := d.DB.Model(new(model.Incident)).Where("application_id = ?", applicationID).Count(&count).Error
	return count, err
}
