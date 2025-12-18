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
	ErrCode   int       `json:"errcode"`
	ErrMsg    string    `json:"errmsg"`
	PhoneInfo PhoneInfo `json:"phone_info"`
}

type PhoneInfo struct {
	PhoneNumber     string    `json:"phoneNumber"`
	PurePhoneNumber string    `json:"purePhoneNumber"`
	CountryCode     string    `json:"countryCode"`
	Watermark       Watermark `json:"watermark"`
}

type Watermark struct {
	Timestamp int64  `json:"timestamp"`
	AppID     string `json:"appid"`
}

type BindPhoneRequest struct {
	PhoneCode string `json:"phone_code"` //与wx.LOGIN的code不一样
}

type WxLoginResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type LoginRep struct {
	Token       string `json:"token"`
	UserId      int    `json:"user_id"`
	OpenId      string `json:"open_id"`
	PhoneNumber string `json:"phone_number"`
}

type BindPhoneRep struct {
	PhoneNumber string `json:"phone_number"`
}
