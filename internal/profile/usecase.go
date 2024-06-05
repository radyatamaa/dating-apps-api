package profile

import (
	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/radyatamaa/dating-apps-api/internal/domain"
)

// UseCase Interface
type UseCase interface {
	GetProfiles(beegoCtx *beegoContext.Context, page, limit, offset int,latitude,longitude string)(*domain.GetProfilesResponsePaginationResponse, error)
	UpdateLiveLocationProfiles(beegoCtx *beegoContext.Context, request domain.UpdateLiveLocationProfilesRequest) error
}
