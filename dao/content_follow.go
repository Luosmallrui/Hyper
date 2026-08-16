package dao

import (
	"Hyper/models"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FollowContent validates a public content target and creates an idempotent
// object-level follow relation. Content follows are intentionally allowed for
// the owner too: an organizer can follow their own activity or venue so it is
// retained in their personal follow list just like any other content.
func FollowContent(ctx context.Context, db *gorm.DB, userID int64, targetType string, targetID int64) error {
	if _, err := ResolveContentFollowOwner(ctx, db, targetType, targetID); err != nil {
		return err
	}
	if userID <= 0 {
		return errors.New("用户未登录")
	}
	return db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&models.ContentFollow{
		UserID: userID, TargetType: targetType, TargetID: targetID,
	}).Error
}

func UnfollowContent(ctx context.Context, db *gorm.DB, userID int64, targetType string, targetID int64) error {
	if !isContentFollowTarget(targetType) || targetID <= 0 || userID <= 0 {
		return fmt.Errorf("关注目标无效")
	}
	return db.WithContext(ctx).
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Delete(&models.ContentFollow{}).Error
}

func ResolveContentFollowOwner(ctx context.Context, db *gorm.DB, targetType string, targetID int64) (int64, error) {
	if targetID <= 0 {
		return 0, fmt.Errorf("关注目标无效")
	}

	var ownerID int64
	var err error
	switch targetType {
	case models.ContentFollowTargetActivity:
		err = db.WithContext(ctx).Table("activities a").
			Select("o.user_id").
			Joins("JOIN organizers o ON o.id = a.organizer_id").
			Where("a.id = ? AND a.type <> ? AND a.status = ? AND a.is_hidden = 0", targetID, models.ActivityTypeVenue, models.ActivityStatusOnline).
			Scan(&ownerID).Error
	case models.ContentFollowTargetVenue:
		err = db.WithContext(ctx).Table("organizers o").
			Select("o.user_id").
			Where("o.id = ? AND o.status = ? AND o.enabled = 1", targetID, models.OrganizerStatusApproved).
			// A venue is now modeled by organizers.type=venue. Keep the legacy
			// activities lookup for migrated records that have not been normalized.
			Where(`o.type = ? OR EXISTS (
				SELECT 1 FROM activities a
				WHERE a.organizer_id = o.id AND a.type = ? AND a.status = ? AND a.is_hidden = 0
			)`, models.OrganizerTypeVenue, models.ActivityTypeVenue, models.ActivityStatusOnline).
			Scan(&ownerID).Error
	case models.ContentFollowTargetOrganizer:
		err = db.WithContext(ctx).Table("organizers o").
			Select("o.user_id").
			Where("o.id = ? AND o.status = ? AND o.enabled = 1", targetID, models.OrganizerStatusApproved).
			Scan(&ownerID).Error
	case models.ContentFollowTargetParty:
		err = db.WithContext(ctx).Table("parties").
			Select("user_id").
			Where("id = ? AND status = ?", targetID, "active").
			Scan(&ownerID).Error
	default:
		return 0, fmt.Errorf("target_type 仅支持 activity、venue、organizer、party")
	}
	if err != nil {
		return 0, err
	}
	if ownerID == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	return ownerID, nil
}

// LoadContentFollowStats returns target fan counts and whether userID follows
// each target. Empty input is handled without querying the database.
func LoadContentFollowStats(ctx context.Context, db *gorm.DB, targetType string, targetIDs []int64, userID int64) (map[int64]int64, map[int64]bool, error) {
	counts := make(map[int64]int64)
	followed := make(map[int64]bool)
	if !isContentFollowTarget(targetType) || len(targetIDs) == 0 {
		return counts, followed, nil
	}

	type countRow struct {
		TargetID int64
		Count    int64
	}
	var rows []countRow
	if err := db.WithContext(ctx).Model(&models.ContentFollow{}).
		Select("target_id, COUNT(*) AS count").
		Where("target_type = ? AND target_id IN ?", targetType, targetIDs).
		Group("target_id").
		Scan(&rows).Error; err != nil {
		return nil, nil, err
	}
	for _, row := range rows {
		counts[row.TargetID] = row.Count
	}
	if userID <= 0 {
		return counts, followed, nil
	}

	var followedIDs []int64
	if err := db.WithContext(ctx).Model(&models.ContentFollow{}).
		Where("user_id = ? AND target_type = ? AND target_id IN ?", userID, targetType, targetIDs).
		Pluck("target_id", &followedIDs).Error; err != nil {
		return nil, nil, err
	}
	for _, id := range followedIDs {
		followed[id] = true
	}
	return counts, followed, nil
}

func isContentFollowTarget(targetType string) bool {
	switch targetType {
	case models.ContentFollowTargetActivity, models.ContentFollowTargetVenue, models.ContentFollowTargetOrganizer, models.ContentFollowTargetParty:
		return true
	default:
		return false
	}
}
