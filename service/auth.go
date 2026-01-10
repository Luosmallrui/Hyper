package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/pkg/encrypt"
	"Hyper/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ IUserService = (*UserService)(nil)

type IUserService interface {
	GetOrCreateByOpenID(ctx context.Context, openid string) (*models.Users, error)
	Register(ctx context.Context, opt *UserRegisterOpt) (*models.Users, error)
	Login(mobile string, password string) (*models.Users, error)
	Forget(opt *UserForgetOpt) (bool, error)
	UpdatePassword(uid int, oldPassword string, password string) error
	UpdateMobile(ctx context.Context, UserId int, PhoneNumber string) error
	Update(ctx context.Context, userID int, req *types.UpdateUserReq) error
	BatchGetUserInfo(ctx context.Context, uids []uint64) map[uint64]UserInfo
	GetUserAvatar(ctx context.Context, uid int64) (string, string, error)
	GetUserInfo(ctx context.Context, uid int) (*models.Users, error)
}

type UserService struct {
	UsersRepo *dao.Users
	Redis     *redis.Client
	DB        *gorm.DB
}

func (s *UserService) GetUserInfo(ctx context.Context, uid int) (*models.Users, error) {
	var users *models.Users
	err := s.DB.WithContext(ctx).Where("id = ?", uid).First(&users).Error
	return users, err
}

func (s *UserService) GetUserAvatar(ctx context.Context, uid int64) (string, string, error) {
	var user types.UserProfile
	err := s.DB.WithContext(ctx).
		Table("users").
		Select("avatar", "nickname").
		Where("id = ?", uid).
		Take(&user).Error
	if err != nil {
		return "", "", err
	}
	return user.Avatar, user.Nickname, nil
}

func (s *UserService) UpdateMobile(ctx context.Context, UserId int, PhoneNumber string) error {
	if PhoneNumber == "" {
		return errors.New("手机号不能为空")
	}

	user, err := s.UsersRepo.FindById(ctx, UserId)
	if err != nil || user.Id == 0 {
		return errors.New("用户不存在")
	}

	err = s.UsersRepo.UpdateById(ctx, int64(user.Id), map[string]any{
		"mobile":     PhoneNumber,
		"updated_at": time.Now(),
	})

	return err
}

func (s *UserService) GetOrCreateByOpenID(ctx context.Context, openid string) (*models.Users, error) {
	if openid == "" {
		return nil, errors.New("openid 不能为空")
	}
	return s.UsersRepo.GetOrCreateByOpenID(ctx, openid)
}

type UserRegisterOpt struct {
	Nickname string `json:"nickname"`
	Mobile   string `json:"mobile"`
	Password string `json:"password"`
	Platform string `json:"platform"`
	Email    string `json:"email"`
}

// Register 注册用户
func (s *UserService) Register(ctx context.Context, opt *UserRegisterOpt) (*models.Users, error) {
	if s.UsersRepo.IsMobileExist(ctx, opt.Mobile) {
		return nil, errors.New("账号已存在! ")
	}

	user := &models.Users{
		Mobile:    opt.Mobile,
		Nickname:  opt.Nickname,
		Avatar:    "",
		Gender:    models.UsersGenderDefault,
		Motto:     "",
		Email:     "",
		Birthday:  "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.UsersRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 登录处理
func (s *UserService) Login(mobile string, password string) (*models.Users, error) {

	user, err := s.UsersRepo.FindByMobile(context.Background(), mobile)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("登录账号不存在! ")
		}

		return nil, err
	}

	if !encrypt.VerifyPassword(user.Password, password) {
		return nil, errors.New("登录密码填写错误! ")
	}

	return user, nil
}

// UserForgetOpt ForgetRequest 账号找回接口验证
type UserForgetOpt struct {
	Mobile   string
	Password string
	SmsCode  string
}

// Forget 账号找回
func (s *UserService) Forget(opt *UserForgetOpt) (bool, error) {

	user, err := s.UsersRepo.FindByMobile(context.Background(), opt.Mobile)
	if err != nil || user.Id == 0 {
		return false, errors.New("账号不存在! ")
	}
	return true, nil
}

// UpdatePassword 修改用户密码
func (s *UserService) UpdatePassword(uid int, oldPassword string, password string) error {

	user, err := s.UsersRepo.FindById(context.TODO(), uid)
	if err != nil {
		return errors.New("用户不存在！")
	}

	if !encrypt.VerifyPassword(user.Password, oldPassword) {
		return errors.New("密码验证不正确！")
	}

	return err
}

func (s *UserService) Update(ctx context.Context, userID int, req *types.UpdateUserReq) error {
	updates := make(map[string]interface{})

	if req.Nickname != nil {
		updates["nickname"] = *req.Nickname
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}
	if req.Motto != nil {
		updates["motto"] = *req.Motto
	}
	if req.Birthday != nil {
		updates["birthday"] = *req.Birthday
	}

	if len(updates) == 0 {
		return nil
	}
	err := s.UsersRepo.Update(ctx, uint(userID), updates)
	if err != nil {
		return fmt.Errorf("db update failed: %w", err)
	}
	return nil
}

type UserInfo struct {
	Avatar   string `json:"avatar"`
	Nickname string `json:"nickname"`
}

func (s *UserService) BatchGetUserInfo(ctx context.Context, uids []uint64) map[uint64]UserInfo {
	result := make(map[uint64]UserInfo)
	if len(uids) == 0 {
		return result
	}

	// 1. Redis MGet 批量获取
	keys := make([]string, len(uids))
	for i, id := range uids {
		keys[i] = fmt.Sprintf("user:info:%d", id)
	}

	cacheRes, _ := s.Redis.MGet(ctx, keys...).Result()

	missingIds := make([]uint64, 0)
	for i, val := range cacheRes {
		if val != nil {
			var info UserInfo
			_ = json.Unmarshal([]byte(val.(string)), &info)
			result[uids[i]] = info
		} else {
			missingIds = append(missingIds, uids[i])
		}
	}

	// 2. 如果有缓存缺失，查数据库
	if len(missingIds) > 0 {
		var dbUsers []struct {
			Id       uint64
			Avatar   string
			Nickname string
		}
		s.DB.Table("users").Where("id IN ?", missingIds).Find(&dbUsers)

		pipe := s.Redis.Pipeline()
		for _, user := range dbUsers {
			info := UserInfo{Avatar: user.Avatar, Nickname: user.Nickname}
			result[user.Id] = info

			// 写入缓存供下次使用
			data, _ := json.Marshal(info)
			pipe.Set(ctx, fmt.Sprintf("user:info:%d", user.Id), data, 15*time.Second)
		}
		_, _ = pipe.Exec(ctx)
	}

	return result
}
