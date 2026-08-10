package models

const (
	ContentTagTypeCoupon = "coupon_tag"

	ContentTagTargetActivity = "activity"
	ContentTagTargetVenue    = "venue"
	ContentTagTargetParty    = "party"
)

type ContentTagItem struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image,omitempty"`
	Value string `json:"value,omitempty"`
	Sort  int    `json:"sort"`
}

// ContentTagRelation binds a configured coupon tag to one business entity.
// target_type + target_id keeps the model extensible without duplicating
// activity, venue and legacy-party relation tables.
type ContentTagRelation struct {
	ID         int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	TagID      int64  `gorm:"column:tag_id;not null;uniqueIndex:uk_content_tag_target,priority:1;index" json:"tag_id"`
	TargetType string `gorm:"column:target_type;size:20;not null;uniqueIndex:uk_content_tag_target,priority:2;index" json:"target_type"`
	TargetID   int64  `gorm:"column:target_id;not null;uniqueIndex:uk_content_tag_target,priority:3;index" json:"target_id"`
}

func (ContentTagRelation) TableName() string { return "content_tag_relations" }
