package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/pkg/encrypt"
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
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
	UpdateMobile(ctx context.Context, UserId int, PhoneNumber string) error
	SendVerifyCode(ctx context.Context, mobile string) error
	UpdateMobileWithSms(ctx context.Context, mobile string, UserId int, inputCode string) error
}

type UserService struct {
	UsersRepo *dao.Users
	Redis     *redis.Client
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

// Modify send sms code 模拟发送短信验证码
func (s *UserService) SendVerifyCode(ctx context.Context, mobile string) error {
	//先定一个验证码，
	code := "222444"
	//存入redis
	err := s.Redis.Set(ctx, "sms:bind"+mobile, code, 5*time.Minute).Err()
	if err != nil {
		return err
	}
	println("模拟发送短信验证码，验证码为：", code)
	return nil
}

func (s *UserService) UpdateMobileWithSms(ctx context.Context, mobile string, UserId int, inputCode string) error {
	//校验
	cachedCode, err := s.Redis.Get(ctx, "sms:bind"+mobile).Result()
	if err != nil {
		errors.New("验证码已过期或着未发送")
	}
	if cachedCode != inputCode {
		return errors.New("验证码错误")
	}

	existUser, err := s.UsersRepo.FindByMobile(ctx, mobile)
	if err == nil {
		if int(existUser.Id) != UserId {
			return errors.New("该手机号已被绑定")
		}
	}
	err = s.UsersRepo.UpdateById(ctx, int64(UserId), map[string]any{
		"mobile":     mobile,
		"updated_at": time.Now(),
	})

	if err == nil {
		s.Redis.Del(ctx, "sms:bind"+mobile)
	}
	return err

}
