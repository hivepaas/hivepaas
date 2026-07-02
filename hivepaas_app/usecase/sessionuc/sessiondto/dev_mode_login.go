package sessiondto

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type DevModeLoginReq struct {
	UserID string `mapstructure:"userId"`
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
