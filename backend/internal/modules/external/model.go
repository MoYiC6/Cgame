package external

import "time"

type UserOAuth struct {
	ID         int64
	UserID     int64
	Platform   string
	OpenID     string
	UnionID    *string
	Nickname   *string
	Avatar     *string
	SessionKey *string
	Phone      *string
	BoundAt    time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type UserToken struct {
	ID           int64
	UserID       int64
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

type ScanLoginSession struct {
	LoginKey   string
	Status     string
	UserID     *int64
	Token      *string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type WechatLoginRequest struct {
	Code     string `json:"code"`
	Platform string `json:"platform"`
	Type     string `json:"type"`
	AppID    string `json:"appId"`
}

type WechatBindRequest struct {
	Platform string `json:"platform"`
	Code     string `json:"code"`
}

type WechatPhoneRequest struct {
	Code string `json:"code"`
}

type WechatPhoneResponse struct {
	PhoneNumber string `json:"phoneNumber"`
	PurePhone   string `json:"purePhone"`
}

type WechatQrCodeLoginResponse struct {
	QrCodeImage string `json:"qrCodeImage"`
	LoginKey    string `json:"loginKey"`
}

type KookBindCodeResponse struct {
	BindCode string `json:"bindCode"`
}

type KookBindingStatusResponse struct {
	Bound      bool   `json:"bound"`
	KookUserID string `json:"kookUserId"`
	KookNickname string `json:"kookNickname"`
}
