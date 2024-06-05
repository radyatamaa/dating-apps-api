package swipe

import (
	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/radyatamaa/dating-apps-api/internal/domain"
)

// UseCase Interface
type UseCase interface {
	SwipeProfile(beegoCtx *beegoContext.Context, request domain.SwipeProfileRequest) error
}