package service

import (
	"Hyper/config"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ ISMSService = (*SMSService)(nil)

type SMSService struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Client
}

type ISMSService interface {
	SendVerificationCode(ctx context.Context, phone string) error
	VerifyCode(ctx context.Context, phone, code string) (bool, error)
}

func (s *SMSService) GenerateCode(length int) string {
	const digits = "0123456789"
	code := make([]byte, length)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		code[i] = digits[n.Int64()]
	}
	return string(code)
}

// SendVerificationCode 发送验证码
func (s *SMSService) SendVerificationCode(ctx context.Context, phone string) error {
	cooldownKey := fmt.Sprintf("sms:cooldown:%s", phone)
	exists, err := s.Redis.Exists(ctx, cooldownKey).Result()
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("请稍后再试，验证码已发送")
	}

	code := s.GenerateCode(6)

	client, err := dysmsapi.NewClientWithAccessKey(
		"cn-chengdu",
		s.Config.Oss.AccessKeyID,
		s.Config.Oss.AccessKeySecret,
	)
	if err != nil {
		return fmt.Errorf("创建SMS客户端失败: %w", err)
	}

	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.PhoneNumbers = phone
	request.SignName = "四川骇铂尔科技"
	request.TemplateCode = "SMS_499105975"
	request.TemplateParam = fmt.Sprintf(`{"code":"%s"}`, code)

	response, err := client.SendSms(request)
	if err != nil {
		return fmt.Errorf("发送短信失败: %w", err)
	}

	if response.Code != "OK" {
		return fmt.Errorf("短信发送失败: %s", response.Message)
	}

	// 4. 存储验证码到Redis
	codeKey := fmt.Sprintf("sms:code:%s", phone)
	err = s.Redis.Set(ctx, codeKey, code, 5*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("存储验证码失败: %w", err)
	}

	// 5. 设置60秒冷却期
	err = s.Redis.Set(ctx, cooldownKey, "1", 60*time.Second).Err()
	if err != nil {
		return fmt.Errorf("设置冷却期失败: %w", err)
	}

	return nil
}

// VerifyCode 验证验证码
func (s *SMSService) VerifyCode(ctx context.Context, phone, code string) (bool, error) {
	codeKey := fmt.Sprintf("sms:code:%s", phone)

	storedCode, err := s.Redis.Get(ctx, codeKey).Result()
	if errors.Is(err, redis.Nil) {
		return false, fmt.Errorf("验证码已过期或不存在")
	}
	if err != nil {
		return false, fmt.Errorf("redis error: %w", err)
	}

	if storedCode != code {
		return false, fmt.Errorf("验证码错误")
	}

	s.Redis.Del(ctx, codeKey)

	return true, nil
}
