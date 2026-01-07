package types

type User struct {
	Mobile   string `json:"mobile"`   // 手机号
	Nickname string `json:"nickname"` // 用户昵称
	Avatar   string `json:"avatar"`   // 用户头像地址
	Gender   int    `json:"gender"`   // 用户性别 1:男 2:女 3:未知
	Password string `json:"password"` // 用户密码
	Motto    string `json:"motto"`    // 用户座右铭
	Email    string `json:"email"`    // 用户邮箱
	Birthday string `json:"birthday"` // 生日
}

type UpdateUserReq struct {
	Nickname *string `json:"nickname"`
	Avatar   *string `json:"avatar"`
	Gender   *int    `json:"gender"`
	Motto    *string `json:"motto"`
	Birthday *string `json:"birthday"`
	// 手机号、邮箱、密码通常通过专门的修改接口（需验证码或旧密码）
}

type UploadAvatarRes struct {
	Url string `json:"url"`
}
