package demohandler

import (
	"example.com/m/hivepaas_app/interface/api/handler/basehandler"
	"example.com/m/hivepaas_app/usecase/demouc/demodto"
)

type Handler struct {
	*basehandler.Handler
}

// ListItems documents search and its paging, and reads status too.
// @Id      listItems
// @Param   search query string false "`search=<text>`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   stale query string false "nothing reads it"
// @Router  /items [get]
func (h *Handler) ListItems(ctx *gin.Context) {
	req := demodto.NewListItemReq()
	if err := h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		return
	}
}

// GetItem has a path parameter it does not document, and one it documents that
// the path does not have; and the id of another handler.
// @Id      listItems
// @Param   other path string true "not in the path"
// @Router  /items/{itemID} [get]
func (h *Handler) GetItem(ctx *gin.Context) {
	h.Get(ctx, basehandler.KindItem)
}

// DownloadItem reads one parameter by name, and ignores one it decodes. It has
// no @Id, and two routes.
// @Param   itemID path string true "item ID"
// @Param   inline query bool false "`inline=true`"
// openapi:ignore-param reveal - nothing here is secret
// @Router  /items/{itemID}/download [get]
// @Router  /items/{itemID}/download [post]
func (h *Handler) DownloadItem(ctx *gin.Context) {
	_ = ctx.Query("inline")
	h.Get(ctx, basehandler.KindItem)
}

// notAHandler has no @Router: it is not checked.
func (h *Handler) notAHandler(ctx *gin.Context) {}
