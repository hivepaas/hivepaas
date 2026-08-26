package imapi

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestSendMessage(t *testing.T) {
	ctx := context.Background()

	t.Run("empty message returns nil", func(t *testing.T) {
		err := SendMessage(ctx, nil, "")
		assert.NoError(t, err)
	})

	t.Run("nil setting returns missing error", func(t *testing.T) {
		err := SendMessage(ctx, nil, "hello")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrMissing))
	})

	t.Run("unsupported setting type", func(t *testing.T) {
		setting := &entity.Setting{
			Type: base.SettingTypeEmail,
		}
		err := SendMessage(ctx, setting, "hello")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrSettingTypeUnsupported))
	})

	t.Run("missing IM service data", func(t *testing.T) {
		setting := &entity.Setting{
			Type: base.SettingTypeIMService,
			Kind: string(base.IMServiceKindSlack),
		}
		err := SendMessage(ctx, setting, "hello")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrMissing))
	})

	t.Run("unsupported IM service kind", func(t *testing.T) {
		setting := &entity.Setting{
			Type: base.SettingTypeIMService,
			Kind: "unsupported_kind",
			Data: "{}",
		}
		err := SendMessage(ctx, setting, "hello")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrIMServiceUnsupported))
	})

	t.Run("missing Slack setting", func(t *testing.T) {
		setting := &entity.Setting{
			Type: base.SettingTypeIMService,
			Kind: string(base.IMServiceKindSlack),
			Data: "{}",
		}
		err := SendMessage(ctx, setting, "hello")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrMissing))
	})

	t.Run("missing Discord setting", func(t *testing.T) {
		setting := &entity.Setting{
			Type: base.SettingTypeIMService,
			Kind: string(base.IMServiceKindDiscord),
			Data: "{}",
		}
		err := SendMessage(ctx, setting, "hello")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrMissing))
	})

	t.Run("missing Telegram setting", func(t *testing.T) {
		setting := &entity.Setting{
			Type: base.SettingTypeIMService,
			Kind: string(base.IMServiceKindTelegram),
			Data: "{}",
		}
		err := SendMessage(ctx, setting, "hello")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrMissing))
	})
}

func TestSendMessageWithRetry(t *testing.T) {
	ctx := context.Background()

	t.Run("empty message returns nil", func(t *testing.T) {
		err := SendMessageWithRetry(ctx, nil, "", 2, time.Second)
		assert.NoError(t, err)
	})

	t.Run("nil setting returns missing error", func(t *testing.T) {
		err := SendMessageWithRetry(ctx, nil, "hello", 2, time.Second)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrMissing))
	})
}
