package usecase

import (
	"context"
	"fmt"
	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/radyatamaa/dating-apps-api/internal/domain"
	"github.com/radyatamaa/dating-apps-api/internal/profile"
	"github.com/radyatamaa/dating-apps-api/internal/swipe"
	"github.com/radyatamaa/dating-apps-api/pkg/database/paginator"
	"github.com/radyatamaa/dating-apps-api/pkg/helper"
	"github.com/radyatamaa/dating-apps-api/pkg/jwt"
	"github.com/radyatamaa/dating-apps-api/pkg/zaplogger"
	"time"
)

type profileUseCase struct {
	zapLogger                  zaplogger.Logger
	contextTimeout             time.Duration
	mysqlProfileRepository    profile.MysqlRepository
	mysqlSwipeRepository    swipe.MysqlRepository
}



func NewProfileUseCase(timeout time.Duration,
	mysqlProfileRepository    profile.MysqlRepository,
	mysqlSwipeRepository    swipe.MysqlRepository,
	zapLogger zaplogger.Logger) profile.UseCase {
	return &profileUseCase{
		mysqlSwipeRepository:mysqlSwipeRepository,
		mysqlProfileRepository:    mysqlProfileRepository,
		contextTimeout:             timeout,
		zapLogger:                  zapLogger,
	}
}

/////////////////// GetProfiles
func (r profileUseCase) fetchProfileWithFilterAndPagination(ctx context.Context, limit, offset int, fields []string,filter []string, order string, args ...interface{}) (*paginator.Paginator, error) {
	var entity []domain.ProfileQueryWithUser
	paging, err := r.mysqlProfileRepository.FetchWithFilterAndPagination(
		ctx,
		limit,
		offset,
		order,
		fields,
		[]string{
			"INNER JOIN users ON users.id = profile.id",
		},
		filter,
		&entity, args...,
	)
	if err != nil {
		return nil, err
	}

	return paging, nil
}
func (r profileUseCase) fetchSwipeWithFilter(ctx context.Context, limit, offset int, filter []string, args ...interface{}) ([]domain.Swipe, error) {

	if data, err := r.mysqlSwipeRepository.FetchWithFilter(
		ctx,
		limit,
		offset,
		"id ASC",
		[]string{
			"*",
		},
		[]string{},
		filter,
		&[]domain.Swipe{}, args); err != nil {
		return nil, err
	} else {
		if result, ok := data.(*[]domain.Swipe); !ok {
			return []domain.Swipe{}, nil
		} else {
			return *result, nil
		}
	}
}
func (p profileUseCase) GetProfiles(beegoCtx *beegoContext.Context, page, limit, offset int,latitude,longitude string) (*domain.GetProfilesResponsePaginationResponse, error) {
	ctx, cancel := context.WithTimeout(beegoCtx.Request.Context(), p.contextTimeout)
	defer cancel()

	userLogin := beegoCtx.Request.Context().Value("JWT_PAYLOAD").(jwt.Payload)

	fetchSwipes, err := p.fetchSwipeWithFilter(ctx, 0, 0,
		[]string{"user_id = ?"},
		userLogin["uid"].(float64))
	if err != nil {
		beegoCtx.Input.SetData("stackTrace", p.zapLogger.SetMessageLog(err))
		return nil, err
	}

	excludeProfileId := []int{int(userLogin["profile_id"].(float64))}
	for i := range fetchSwipes {
		if fetchSwipes[i].UpdatedAt.Format(helper.DateFormatDefault) == time.Now().Format(helper.DateFormatDefault) ||
			fetchSwipes[i].SwipeType == "LIKE"{
			excludeProfileId = append(excludeProfileId,fetchSwipes[i].ProfileID)
		}
	}

	filters := make([]string,0)
	args := make([]interface{}, 0)
	fields := []string{
		"profile.*",
		"users.premium_expires_at",
	}
	order := "RAND()"
	if len(excludeProfileId) > 0 {
		filters = append(filters,"profile.id not in (?)")
		args = append(args,excludeProfileId)
	}

	if latitude != "" && longitude != "" {
		fields = append(fields,fmt.Sprintf(`(6371 * 
 		acos(cos(radians(%s)) * 
		cos(radians(latitude)) * 
		cos(radians(longitude) - 
		radians(%s)) + 
		sin(radians(%s)) * sin(radians(latitude)))) AS distance`,latitude,longitude,latitude))
		order = "distance"
	}

	fetchProfiles, err := p.fetchProfileWithFilterAndPagination(ctx, limit, offset,fields,filters , order, args...)
	if err != nil {
		beegoCtx.Input.SetData("stackTrace", p.zapLogger.SetMessageLog(err))
		return nil, err
	}

	datas := make([]domain.GetProfilesResponse, 0)
	records := fetchProfiles.Records.(*[]domain.ProfileQueryWithUser)
	if records != nil {
		for _, e := range *records {
			datas = append(datas, domain.FromProfileToGetProfilesResponse(e))
		}
	}

	result := domain.ToGetProfilesResponsePaginationResponsee(datas, page, limit, offset, int(fetchProfiles.Total))

	return result,nil
}
//////////////////

func (r profileUseCase) UpdateLiveLocationProfiles(beegoCtx *beegoContext.Context, request domain.UpdateLiveLocationProfilesRequest) error {
	ctx, cancel := context.WithTimeout(beegoCtx.Request.Context(), r.contextTimeout)
	defer cancel()

	userLogin := beegoCtx.Request.Context().Value("JWT_PAYLOAD").(jwt.Payload)

	err := r.mysqlProfileRepository.UpdateSelectedField(ctx,[]string{
		"longitude",
		"latitude",
		"updated_at",
	}, map[string]interface{}{
		"longitude": request.Longitude,
		"latitude": request.Latitude,
		"updated_at": time.Now(),
	},int(userLogin["uid"].(float64)))
	if err != nil {
		beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
		return err
	}

	return nil
}