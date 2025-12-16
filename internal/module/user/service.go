package user

import "context"

// Service 接口
type Service interface {
	GetUser(ctx context.Context, id uint) (*UserModel, error)
	CreateUser(ctx context.Context, id uint) (*UserModel, error)
}

// service 实现
type service struct {
	repo Repository
}

// NewService 构造函数
// 参数 Repository 会由 Wire 自动从 NewRepository() 注入进来
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetUser(ctx context.Context, id uint) (*UserModel, error) {
	// 这里可以写业务逻辑，比如判断用户状态等
	return s.repo.FindByID(ctx, id)
}

func (s *service) CreateUser(ctx context.Context, id uint) (*UserModel, error) {
	// 这里可以写业务逻辑，比如判断用户状态等
	return s.repo.FindByID(ctx, id)
}
