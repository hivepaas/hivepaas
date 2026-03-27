package sessiondto

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type DevModeLoginReq struct {
	UserID string `mapstructure:"userID"`
}

func NewDevModeLoginReq() *DevModeLoginReq {
	return &DevModeLoginReq{}
}

type DevModeLoginResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *DevModeLoginDataResp `json:"data"`
}

type DevModeLoginDataResp struct {
	AccessToken     string    `json:"accessToken"`
	AccessTokenExp  time.Time `json:"accessTokenExp"`
	RefreshToken    string    `json:"refreshToken"`
	RefreshTokenExp time.Time `json:"refreshTokenExp"`
}
