package basehandler

import (
	"example.com/m/hivepaas_app/usecase/demouc/demodto"
)

type Handler struct{}

// Get reads the request of the kind it is given, as the settings handlers do.
func (h *Handler) Get(ctx *gin.Context, kind string) {
	var req any
	switch kind {
	case KindItem:
		r := demodto.NewGetItemReq()
		req = r
	}
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		return
	}
}

const KindItem = "item"
