package dao

import (
	"Hyper/models"
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

func ParseContentTagIDs(raw string) ([]int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []int64{}, nil
	}
	seen := make(map[int64]struct{})
	ids := make([]int64, 0)
	for _, value := range strings.Split(raw, ",") {
		id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("tag_ids 必须是正整数列表")
		}
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, nil
}

func ListActiveCouponTags(ctx context.Context, db *gorm.DB) ([]models.AdminCategory, error) {
	var tags []models.AdminCategory
	err := db.WithContext(ctx).
		Where("type = ? AND status = ?", models.ContentTagTypeCoupon, 1).
		Order("sort ASC, id ASC").
		Find(&tags).Error
	return tags, err
}

func ValidateActiveCouponTagIDs(ctx context.Context, db *gorm.DB, tagIDs []int64) ([]int64, error) {
	unique := make(map[int64]struct{}, len(tagIDs))
	ids := make([]int64, 0, len(tagIDs))
	for _, id := range tagIDs {
		if id <= 0 {
			return nil, fmt.Errorf("标签 ID 必须为正整数")
		}
		if _, ok := unique[id]; !ok {
			unique[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return ids, nil
	}
	var count int64
	if err := db.WithContext(ctx).Model(&models.AdminCategory{}).
		Where("type = ? AND status = ? AND id IN ?", models.ContentTagTypeCoupon, 1, ids).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count != int64(len(ids)) {
		return nil, fmt.Errorf("标签不存在或已停用")
	}
	return ids, nil
}

func ReplaceContentTags(ctx context.Context, db *gorm.DB, targetType string, targetID int64, tagIDs []int64) error {
	if !isContentTagTarget(targetType) || targetID <= 0 {
		return fmt.Errorf("标签绑定目标无效")
	}
	ids, err := ValidateActiveCouponTagIDs(ctx, db, tagIDs)
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("target_type = ? AND target_id = ?", targetType, targetID).Delete(&models.ContentTagRelation{}).Error; err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		rows := make([]models.ContentTagRelation, 0, len(ids))
		for _, tagID := range ids {
			rows = append(rows, models.ContentTagRelation{TagID: tagID, TargetType: targetType, TargetID: targetID})
		}
		return tx.Create(&rows).Error
	})
}

func LoadContentTags(ctx context.Context, db *gorm.DB, targetType string, targetIDs []int64, includeInactive bool) (map[int64][]models.AdminCategory, error) {
	result := make(map[int64][]models.AdminCategory)
	if len(targetIDs) == 0 {
		return result, nil
	}
	type row struct {
		TargetID int64
		ID       int64
		Type     string
		Name     string
		Image    string
		Value    string
		Sort     int
		Status   int8
	}
	var rows []row
	query := db.WithContext(ctx).Table("content_tag_relations ctr").
		Select("ctr.target_id, ac.id, ac.type, ac.name, ac.image, ac.value, ac.sort, ac.status").
		Joins("JOIN admin_categories ac ON ac.id = ctr.tag_id").
		Where("ctr.target_type = ? AND ctr.target_id IN ? AND ac.type = ?", targetType, targetIDs, models.ContentTagTypeCoupon)
	if !includeInactive {
		query = query.Where("ac.status = ?", 1)
	}
	if err := query.Order("ac.sort ASC, ac.id ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, item := range rows {
		result[item.TargetID] = append(result[item.TargetID], models.AdminCategory{
			ID: item.ID, Type: item.Type, Name: item.Name, Image: item.Image, Value: item.Value, Sort: item.Sort, Status: item.Status,
		})
	}
	return result, nil
}

// ApplyContentTagFilter requires a target to contain every selected active tag.
// outerIDColumn must be a trusted SQL identifier controlled by backend code.
func ApplyContentTagFilter(query *gorm.DB, targetType, outerIDColumn string, tagIDs []int64) *gorm.DB {
	if len(tagIDs) == 0 {
		return query
	}
	condition, args := ContentTagMatchSQL(targetType, outerIDColumn, tagIDs)
	return query.Where(condition, args...)
}

func ContentTagMatchSQL(targetType, outerIDColumn string, tagIDs []int64) (string, []any) {
	return `EXISTS (
		SELECT 1
		FROM content_tag_relations ctr
		JOIN admin_categories ac ON ac.id = ctr.tag_id
		WHERE ctr.target_type = ?
		  AND ctr.target_id = ` + outerIDColumn + `
		  AND ctr.tag_id IN ?
		  AND ac.type = ?
		  AND ac.status = 1
		GROUP BY ctr.target_id
		HAVING COUNT(DISTINCT ctr.tag_id) = ?
	)`, []any{targetType, tagIDs, models.ContentTagTypeCoupon, len(tagIDs)}
}

func isContentTagTarget(targetType string) bool {
	switch targetType {
	case models.ContentTagTargetActivity, models.ContentTagTargetVenue, models.ContentTagTargetParty:
		return true
	default:
		return false
	}
}
