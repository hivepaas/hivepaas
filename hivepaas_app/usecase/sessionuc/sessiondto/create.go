package sessiondto

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type BaseCreateSessionReq struct {
	User         *entity.User
	IsAPIKey     bool
	AccessAction *base.AccessActions

	// Method is how the caller proved who they are, and it is what the session's
	// audit entry is filed under.
	//
	// Required, and refused when empty rather than defaulted. Every way into
	// HivePaaS ends at createSession, so that refusal is what stops a login path
	// added later from being recorded as nothing in particular - or, worse, from
	// being the one door with no record at all.
	Method string
}

type BaseCreateSessionResp struct {
	AccessToken     string    `json:"accessToken"`
	AccessTokenExp  time.Time `json:"accessTokenExp"`
	RefreshToken    string    `json:"refreshToken"`
	RefreshTokenExp time.Time `json:"refreshTokenExp"`
}
