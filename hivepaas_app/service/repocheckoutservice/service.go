package repocheckoutservice

import (
	"context"
)

type Service interface {
	Checkout(ctx context.Context, req *RepoCheckoutReq) (*RepoCheckoutResp, error)
}
