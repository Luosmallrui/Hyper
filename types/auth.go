package types

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type UserResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

type WxLoginRequest struct {
	LoginCode string `json:"code"` // wx.login 获取的 code (用于换 openid)
	//PhoneCode string `json:"phone_code"` // getPhoneNumber 获取的 code
}

type WxSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type UpdateProfileRequest struct {
	AvatarUrl string `json:"avatar_url"`
	Nickname  string `json:"nickname"`
}

type WxPhoneResponse struct {
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	PhoneInfo struct {
		PhoneNumber     string `json:"phoneNumber"`
		PurePhoneNumber string `json:"purePhoneNumber"`
		CountryCode     string `json:"countryCode"`
	} `json:"phone_info"`
}
type BindPhoneRequest struct {
	PhoneCode string `json:"phone_code"`
}

type WxLoginResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}
