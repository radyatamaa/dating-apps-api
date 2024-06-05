package user

import (
	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/radyatamaa/dating-apps-api/internal/domain"
)

// UseCase Interface
type UseCase interface {
	Login(beegoCtx *beegoContext.Context, request domain.LoginRequest)(*domain.LoginResponse, error)
	Register(beegoCtx *beegoContext.Context, request domain.RegisterRequest) error
	PurchasePremiumUpdateStatus(beegoCtx *beegoContext.Context) error
}