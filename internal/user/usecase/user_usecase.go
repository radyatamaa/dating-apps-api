package usecase

import (
	"context"
	"github.com/radyatamaa/dating-apps-api/internal/profile"
	"github.com/radyatamaa/dating-apps-api/internal/user"
	"time"

	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/radyatamaa/dating-apps-api/internal/domain"
	"github.com/radyatamaa/dating-apps-api/pkg/jwt"
	"github.com/radyatamaa/dating-apps-api/pkg/response"
	"github.com/radyatamaa/dating-apps-api/pkg/zaplogger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type userUseCase struct {
	zapLogger                  zaplogger.Logger
	jwtAuth                    jwt.JWT
	expireToken                int
	contextTimeout             time.Duration
	mysqlUserRepository    user.MysqlRepository
	mysqlProfileRepository profile.MysqlRepository
}



func NewUserUseCase(timeout time.Duration,
	mysqlUserRepository    user.MysqlRepository,
	mysqlProfileRepository profile.MysqlRepository,
	jwtAuth jwt.JWT,
	expireToken int,
	zapLogger zaplogger.Logger) user.UseCase {
	return &userUseCase{
		mysqlUserRepository:    mysqlUserRepository,
		mysqlProfileRepository: mysqlProfileRepository,
		contextTimeout:             timeout,
		zapLogger:                  zapLogger,
		jwtAuth:                    jwtAuth,
		expireToken:                expireToken,
	}
}

/////////////////// Login
func (a userUseCase) singleUserWithFilter(ctx context.Context, filter []string, args ...interface{}) (*domain.UserQueryWithProfile, error) {
	var entity domain.UserQueryWithProfile
	if err := a.mysqlUserRepository.SingleWithFilter(
		ctx,
		[]string{
			"users.*",
			"profile.id as profile_id",
		},
		[]string{
			"INNER JOIN profile ON profile.user_id = users.id",
		},
		filter,
		&entity, args...); err != nil {
		return nil, err
	}
	return &entity, nil
}
func (a userUseCase) Login(beegoCtx *beegoContext.Context, request domain.LoginRequest) (*domain.LoginResponse, error) {
	ctx, cancel := context.WithTimeout(beegoCtx.Request.Context(), a.contextTimeout)
	defer cancel()

	res := new(domain.LoginResponse)

	userSingle, err := a.singleUserWithFilter(ctx, []string{"email = ?"}, request.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			beegoCtx.Input.SetData("stackTrace", a.zapLogger.SetMessageLog(response.ErrInvalidEmailPassword))
			return nil, response.ErrInvalidEmailPassword
		}
		beegoCtx.Input.SetData("stackTrace", a.zapLogger.SetMessageLog(err))
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userSingle.PasswordHash), []byte(request.Password)); err != nil {
		beegoCtx.Input.SetData("stackTrace", a.zapLogger.SetMessageLog(response.ErrInvalidEmailPassword))
		return nil, response.ErrInvalidEmailPassword
	}

	token, err := a.jwtAuth.Ctx(ctx).GenerateToken(jwt.Payload{"uid": userSingle.ID, "email": userSingle.Email, "profile_id": userSingle.ProfileId}, beegoCtx.Request.Host, a.expireToken)
	if err != nil {
		beegoCtx.Input.SetData("stackTrace", a.zapLogger.SetMessageLog(err))
		return nil, err
	}

	res.Token = token.Token
	res.ExpiredAt = token.ExpiredAt.String()
	res.User = domain.FromUserToUserLogin(userSingle)

	return res, nil
}
//////////////////

func (r userUseCase) Register(beegoCtx *beegoContext.Context, request domain.RegisterRequest) error {
	ctx, cancel := context.WithTimeout(beegoCtx.Request.Context(), r.contextTimeout)
	defer cancel()

	if err := r.mysqlUserRepository.DB().Transaction(func(tx *gorm.DB) error {
		userId,err := r.mysqlUserRepository.StoreWithTx(ctx,tx,request.ToUser())
		if err != nil {
			beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
			return err
		}

		_,err = r.mysqlProfileRepository.StoreWithTx(ctx,tx,request.ToProfile(userId))
		if err != nil {
			beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (r userUseCase) PurchasePremiumUpdateStatus(beegoCtx *beegoContext.Context) error {
	ctx, cancel := context.WithTimeout(beegoCtx.Request.Context(), r.contextTimeout)
	defer cancel()

	userLogin := beegoCtx.Request.Context().Value("JWT_PAYLOAD").(jwt.Payload)

	userSingle, err := r.singleUserWithFilter(ctx, []string{"id = ?"}, userLogin["uid"].(float64))
	if err != nil {
		beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
		return err
	}

	if err = r.mysqlUserRepository.DB().Transaction(func(tx *gorm.DB) error {
		err = r.mysqlUserRepository.UpdateSelectedFieldWithTx(ctx,tx,[]string{}, map[string]interface{}{
			"premium_expires_at" : time.Now().AddDate(0,1,0),
			"updated_at" : time.Now(),
		},userSingle.ID)
		if err != nil {
			beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
			return err
		}

		err = r.mysqlProfileRepository.UpdateSelectedFieldWithTx(ctx,tx,[]string{}, map[string]interface{}{
			"verified" : true,
			"updated_at" : time.Now(),
		},userSingle.ID)
		if err != nil {
			beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
			return err
		}

		return nil
	}); err != nil {
		return err
	}


	return nil
}