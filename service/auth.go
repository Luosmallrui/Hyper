package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/pkg/encrypt"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

var _ IUserService = (*UserService)(nil)

type IUserService interface {
	GetOrCreateByOpenID(ctx context.Context, openid string) (*models.Users, error)
	Register(ctx context.Context, opt *UserRegisterOpt) (*models.Users, error)
	Login(mobile string, password string) (*models.Users, error)
	Forget(opt *UserForgetOpt) (bool, error)
	UpdatePassword(uid int, oldPassword string, password string) error
}

type UserService struct {
	UsersRepo *dao.Users
}

func (s *UserService) GetOrCreateByOpenID(ctx context.Context, openid string) (*models.Users, error) {
	if openid == "" {
		return nil, errors.New("openid 不能为空")
	}
	user, err := s.UsersRepo.FindByWhere(ctx, "open_id = ?", openid)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	user = &models.Users{
		OpenID: openid,
	}
	err = s.UsersRepo.Create(ctx, user)
	if err == nil {
		return user, nil
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		err = s.UsersRepo.Db.Where("open_id = ?", openid).First(&user).Error
		if err == nil {
			return user, nil
		}
	}
	return nil, err

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

	affected, err := s.UsersRepo.UpdateById(context.TODO(), user.Id, map[string]any{
		"password": encrypt.HashPassword(opt.Password),
	})

	return affected > 0, err
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

	_, err = s.UsersRepo.UpdateById(context.TODO(), user.Id, map[string]any{
		"password": encrypt.HashPassword(password),
	})

	return err
}
