package dao

import (
	"Hyper/dao/cache"
	"Hyper/models"
	"context"

	"gorm.io/gorm"
)

type GroupMember struct {
	Repo[models.GroupMember]
	relation *cache.Relation
}

func NewGroupMember(db *gorm.DB, relation *cache.Relation) *GroupMember {
	return &GroupMember{Repo: NewRepo[models.GroupMember](db), relation: relation}
}

// IsMaster 判断是否是群主
func (g *GroupMember) IsMaster(ctx context.Context, gid, uid int) bool {
	exist, err := g.Repo.IsExist(ctx, "group_id = ? and user_id = ? and role = ? and is_quit = 0", gid, uid, models.GroupMemberLeaderOwner)
	return err == nil && exist
}

// IsLeader 判断是否是群主或管理员
func (g *GroupMember) IsLeader(ctx context.Context, gid, uid int) bool {
	exist, err := g.Repo.IsExist(ctx, "group_id = ? and user_id = ? and role in ? and is_quit = 0", gid, uid, []int{models.GroupMemberLeaderAdmin, models.GroupMemberLeaderOwner})
	return err == nil && exist
}

// IsMember 检测是属于群成员
func (g *GroupMember) IsMember(ctx context.Context, gid, uid int, cache bool) bool {

	if cache && g.relation.IsGroupRelation(ctx, uid, gid) == nil {
		return true
	}

	exist, err := g.Repo.IsExist(ctx, "group_id = ? and user_id = ? and is_quit = 0", gid, uid)
	if err != nil {
		return false
	}

	if exist {
		g.relation.SetGroupRelation(ctx, uid, gid)
	}

	return exist
}

func (g *GroupMember) FindByUserId(ctx context.Context, gid, uid int) (*models.GroupMember, error) {
	member := &models.GroupMember{}
	err := g.Repo.Model(ctx).Where("group_id = ? and user_id = ?", gid, uid).First(member).Error
	return member, err
}

// GetMemberIds 获取所有群成员用户ID（只查未退群成员）
func (g *GroupMember) GetMemberIds(ctx context.Context, groupId int) ([]int, error) {
	var ids []int
	err := g.Repo.Model(ctx).
		Where("group_id = ? AND is_quit = 0", groupId).
		Pluck("user_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// GetUserGroupIds 获取用户加入的群ID（只查未退群）
func (g *GroupMember) GetUserGroupIds(ctx context.Context, uid int) ([]int, error) {
	var ids []int
	err := g.Repo.Model(ctx).
		Where("user_id = ? AND is_quit = 0", uid).
		Pluck("group_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// CountMemberTotal 统计群成员总数
func (g *GroupMember) CountMemberTotal(ctx context.Context, gid int) int64 {
	count, _ := g.Repo.FindCount(ctx, "group_id = ? and is_quit = 0", gid)
	return count
}

// GetMemberRemark 获取指定群成员的备注信息
func (g *GroupMember) GetMemberRemark(ctx context.Context, groupId int, userId int) string {

	var remarks string
	g.Repo.Model(ctx).Select("user_card").Where("group_id = ? and user_id = ?", groupId, userId).Scan(&remarks)

	return remarks
}

// GetMembers 获取群组成员列表
func (g *GroupMember) GetMembers(ctx context.Context, groupId int) []*models.MemberItem {
	fields := []string{
		"group_member.id",
		"group_member.role",
		"group_member.user_card",
		"group_member.user_id",
		"group_member.is_mute",
		"users.avatar",
		"users.nickname",
		"users.gender",
		"users.motto",
	}

	tx := g.Repo.Db.WithContext(ctx).Table("group_member")
	tx.Joins("left join users on users.id = group_member.user_id")
	tx.Where("group_member.group_id = ? and group_member.is_quit = 0", groupId)
	tx.Order("group_member.role asc") // 群主(1)在前

	var items []*models.MemberItem
	tx.Unscoped().Select(fields).Scan(&items)

	return items
}

type CountGroupMember struct {
	GroupId int `gorm:"column:group_id;"`
	Count   int `gorm:"column:count;"`
}

func (g *GroupMember) CountGroupMemberNum(ids []int) ([]*CountGroupMember, error) {

	var items []*CountGroupMember
	err := g.Repo.Model(context.TODO()).Select("group_id,count(*) as count").Where("group_id in ? and is_quit = 0", ids).Group("group_id").Scan(&items).Error
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (g *GroupMember) CheckUserGroup(ids []int, userId int) ([]int, error) {
	items := make([]int, 0)

	err := g.Repo.Model(context.TODO()).Select("group_id").Where("group_id in ? and user_id = ? and is_quit = 0", ids, userId).Scan(&items).Error
	if err != nil {
		return nil, err
	}

	return items, nil
}
