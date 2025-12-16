package user

import "github.com/google/wire"

// ProviderSet 把这一层的构造函数都暴露出去
// 这样在 main.go 或 wire.go 里只需要引用这个 Set 即可
var ProviderSet = wire.NewSet(
	NewRepository,
	NewService,
	NewHandler,
)
