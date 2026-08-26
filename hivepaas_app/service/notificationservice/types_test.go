package notificationservice

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTaskResultNotificationResp_HasSend(t *testing.T) {
	t.Run("nil or empty map", func(t *testing.T) {
		resp := &TaskResultNotificationResp{
			SendTs: time.Now(),
		}
		assert.False(t, resp.HasSend())

		resp.DeliveryMap = make(map[string]bool)
		assert.False(t, resp.HasSend())
	})

	t.Run("all deliveries failed", func(t *testing.T) {
		resp := &TaskResultNotificationResp{
			SendTs: time.Now(),
			DeliveryMap: map[string]bool{
				"email": false,
				"slack": false,
			},
		}
		assert.False(t, resp.HasSend())
	})

	t.Run("at least one delivery succeeded", func(t *testing.T) {
		resp := &TaskResultNotificationResp{
			SendTs: time.Now(),
			DeliveryMap: map[string]bool{
				"email": false,
				"slack": true,
			},
		}
		assert.True(t, resp.HasSend())
	})
}
