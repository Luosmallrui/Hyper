package service

import (
	"Hyper/models"
	"context"

	"gorm.io/gorm"
)

// resolveOrganizerFollowTarget keeps migrated and legacy venues on the same
// follow relation as map markers. Old venue activities can belong to a
// merchant-type organizer until its data migration is complete.
func resolveOrganizerFollowTarget(ctx context.Context, db *gorm.DB, organizer models.Organizer) (string, error) {
	if organizer.Type == models.OrganizerTypeVenue {
		return models.ContentFollowTargetVenue, nil
	}
	var legacyVenueCount int64
	if err := db.WithContext(ctx).Model(&models.Activity{}).
		Where("organizer_id = ? AND type = ? AND status = ? AND is_hidden = 0", organizer.ID, models.ActivityTypeVenue, models.ActivityStatusOnline).
		Count(&legacyVenueCount).Error; err != nil {
		return "", err
	}
	if legacyVenueCount > 0 {
		return models.ContentFollowTargetVenue, nil
	}
	return models.ContentFollowTargetOrganizer, nil
}
