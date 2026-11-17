package database

import (
	"errors"

	"github.com/gotify/server/v3/model"
	"gorm.io/gorm"
)

// GetPresence returns the stored presence of a user, or nil when it was never reported.
func (d *GormDatabase) GetPresence(userID uint) (*model.Presence, error) {
	presence := new(model.Presence)
	err := d.DB.Where("user_id = ?", userID).First(presence).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return presence, nil
}

// SavePresence inserts or replaces the presence row of a user.
func (d *GormDatabase) SavePresence(presence *model.Presence) error {
	return d.DB.Save(presence).Error
}

// ListPresence returns the presence rows of every user.
func (d *GormDatabase) ListPresence() ([]*model.Presence, error) {
	var presences []*model.Presence
	err := d.DB.Order("user_id ASC").Find(&presences).Error
	return presences, err
}
