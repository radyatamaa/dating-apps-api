package usecase

import (
	"context"
	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/radyatamaa/dating-apps-api/internal/domain"
	"github.com/radyatamaa/dating-apps-api/internal/swipe"
	"github.com/radyatamaa/dating-apps-api/internal/user"
	"github.com/radyatamaa/dating-apps-api/pkg/database/paginator"
	"github.com/radyatamaa/dating-apps-api/pkg/helper"
	"github.com/radyatamaa/dating-apps-api/pkg/jwt"
	"github.com/radyatamaa/dating-apps-api/pkg/response"
	"github.com/radyatamaa/dating-apps-api/pkg/zaplogger"
	"time"
)

type swipeUseCase struct {
	zapLogger                  zaplogger.Logger
	contextTimeout             time.Duration
	mysqlSwipeRepository    swipe.MysqlRepository
	mysqlUserRepository    user.MysqlRepository
}

func NewSwipeUseCase(timeout time.Duration,
	mysqlSwipeRepository    swipe.MysqlRepository,
	mysqlUserRepository    user.MysqlRepository,
	zapLogger zaplogger.Logger) swipe.UseCase {
	return &swipeUseCase{
		mysqlSwipeRepository:    mysqlSwipeRepository,
		mysqlUserRepository:mysqlUserRepository,
		contextTimeout:             timeout,
		zapLogger:                  zapLogger,
	}
}
/////////////////// SwipeProfile
func (a swipeUseCase) singleUserWithFilter(ctx context.Context, filter []string, args ...interface{}) (*domain.User, error) {
	var entity domain.User
	if err := a.mysqlUserRepository.SingleWithFilter(
		ctx,
		[]string{
			"*",
		},
		[]string{},
		filter,
		&entity, args...); err != nil {
		return nil, err
	}
	return &entity, nil
}
func (r swipeUseCase) fetchSwipeWithFilterAndPagination(ctx context.Context, limit, offset int, filter []string, order string, args ...interface{}) (*paginator.Paginator, error) {
	var entity []domain.Swipe
	paging, err := r.mysqlSwipeRepository.FetchWithFilterAndPagination(
		ctx,
		limit,
		offset,
		order,
		[]string{
			"*",
		},
		[]string{},
		filter,
		&entity, args...,
	)
	if err != nil {
		return nil, err
	}

	return paging, nil
}
func (s swipeUseCase) checkDailySwipeQuota(beegoCtx *beegoContext.Context,userId int) (bool,error) {
	fetchSwipes, err := s.fetchSwipeWithFilterAndPagination(beegoCtx.Request.Context(), 1, 0,
		[]string{"user_id = ?","DATE(updated_at) = DATE(?)"},
		"id ASC",
		userId,
		time.Now().Format(helper.DateFormatDefault))
	if err != nil {
		beegoCtx.Input.SetData("stackTrace", s.zapLogger.SetMessageLog(err))
		return false, err
	}

	if  fetchSwipes.Total >= 10 {
		return true, nil
	}

	return false,nil


}
func (a swipeUseCase) singleProfileWithFilter(ctx context.Context, filter []string, args ...interface{}) (*domain.Profile, error) {
	var entity domain.Profile
	if err := a.mysqlUserRepository.SingleWithFilter(
		ctx,
		[]string{
			"*",
		},
		[]string{},
		filter,
		&entity, args...); err != nil {
		return nil, err
	}
	return &entity, nil
}
func (s swipeUseCase) SwipeProfile(beegoCtx *beegoContext.Context, request domain.SwipeProfileRequest) error {
	ctx, cancel := context.WithTimeout(beegoCtx.Request.Context(), s.contextTimeout)
	defer cancel()
	beegoCtx.Request.WithContext(ctx)

	userLogin := beegoCtx.Request.Context().Value("JWT_PAYLOAD").(jwt.Payload)

	userSingle, err := s.singleUserWithFilter(ctx, []string{"id = ?"}, userLogin["uid"].(float64))
	if err != nil {
		beegoCtx.Input.SetData("stackTrace", s.zapLogger.SetMessageLog(err))
		return err
	}

	if !domain.IsPremium(userSingle.PremiumExpiresAt) {
		checkDailySwipeQuota,err := s.checkDailySwipeQuota(beegoCtx,userSingle.ID)
		if err != nil {
			return err
		}

		if checkDailySwipeQuota {
			beegoCtx.Input.SetData("stackTrace", s.zapLogger.SetMessageLog(response.ErrLimitSwipeOrLike))
			return response.ErrLimitSwipeOrLike
		}
	}

	if err = s.mysqlSwipeRepository.Upsert(ctx, []string{"user_id", "profile_id"}, []domain.Swipe{request.ToSwipe(userSingle.ID)}...); err != nil {
		beegoCtx.Input.SetData("stackTrace", s.zapLogger.SetMessageLog(err))
		return err
	}

	return nil
}
//////////////////