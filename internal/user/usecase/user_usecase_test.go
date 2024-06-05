package usecase

import (
	"context"
	"errors"
	"fmt"
	beegoContext "github.com/beego/beego/v2/server/web/context"
	beegoMock "github.com/beego/beego/v2/server/web/mock"
	"github.com/golang/mock/gomock"
	"github.com/radyatamaa/dating-apps-api/internal/domain/mocks"
	"github.com/radyatamaa/dating-apps-api/internal/profile"
	"github.com/radyatamaa/dating-apps-api/internal/user"
	"github.com/radyatamaa/dating-apps-api/pkg/jwt"
	mockJwt "github.com/radyatamaa/dating-apps-api/pkg/jwt/mocks"
	"github.com/radyatamaa/dating-apps-api/pkg/zaplogger"
	mockZaplogger "github.com/radyatamaa/dating-apps-api/pkg/zaplogger/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"
)

type UserUseCaseTestSuite struct {
	suite.Suite
}

func (t *UserUseCaseTestSuite) SetupSuite() {
}

type fields struct {
	zapLogger                 *mockZaplogger.MockLogger
	jwtAuth                   *mockJwt.MockJWT
	expireToken                int
	contextTimeout             time.Duration
	mysqlUserRepository    *mocks.UserMysqlRepository
	mysqlProfileRepository *mocks.ProfileMysqlRepository
}

func toField(ctrl *gomock.Controller) fields {
	return fields{
		zapLogger:                        mockZaplogger.NewMockLogger(ctrl),
		jwtAuth: 						  mockJwt.NewMockJWT(ctrl),
		expireToken: 					  86400,
		contextTimeout:                   time.Second * 30,
		mysqlUserRepository:              mocks.NewUserMysqlRepository(ctrl),
		mysqlProfileRepository:      	  mocks.NewProfileMysqlRepository(ctrl),
	}
}

func (t *UserUseCaseTestSuite) TestNewCustomerUseCase() {
	ctrl := gomock.NewController(t.T())
	defer ctrl.Finish()

	type args struct {
		zapLogger                  zaplogger.Logger
		jwtAuth                    jwt.JWT
		expireToken                int
		contextTimeout             time.Duration
		mysqlUserRepository    user.MysqlRepository
		mysqlProfileRepository profile.MysqlRepository
	}
	tests := []struct {
		name string
		args args
		want user.UseCase
	}{
		{
			name: "success",
			args: args{
				zapLogger:                        mockZaplogger.NewMockLogger(ctrl),
				jwtAuth: 						  mockJwt.NewMockJWT(ctrl),
				expireToken: 					  86400,
				contextTimeout:                   time.Second * 30,
				mysqlUserRepository:              mocks.NewUserMysqlRepository(ctrl),
				mysqlProfileRepository:      	  mocks.NewProfileMysqlRepository(ctrl),
			},
			want: NewUserUseCase(time.Second * 30,mocks.NewUserMysqlRepository(ctrl),mocks.NewProfileMysqlRepository(ctrl),mockJwt.NewMockJWT(ctrl),86400,mockZaplogger.NewMockLogger(ctrl)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func() {
			if got := NewUserUseCase(tt.args.contextTimeout,tt.args.mysqlUserRepository,tt.args.mysqlProfileRepository,tt.args.jwtAuth,tt.args.expireToken,tt.args.zapLogger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf(errors.New("failed"), "NewUserUseCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func (t *UserUseCaseTestSuite) TestUserUseCase_PurchasePremiumUpdateStatus() {
	mockUserLogin := jwt.Payload{"uid": float64(1), "email": "test@gmail.com", "profile_id": float64(1)}
	req := http.Request{}
	req.WithContext(context.Background())
	contextBeego, _ := beegoMock.NewMockContext(&req)
	ctx := context.TODO()
	ctx = context.WithValue(ctx, "JWT_PAYLOAD", mockUserLogin)
	uri := url.URL{
		Scheme: "http",
		Host:   "localhost:8080",
		Path:   "/api/v1/user/purchase-premium",
	}
	contextBeego.Request = httptest.NewRequest(http.MethodPost,uri.String(),nil).WithContext(ctx)

	type args struct {
		beegoCtx   *beegoContext.Context
		token      string
		userId int
		fields []string
		values map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  func(args *args,ctrl *gomock.Controller) fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "success all true",
			wantErr: assert.NoError,
			fields: func(args *args,ctrl *gomock.Controller) fields {
				fields := toField(ctrl)
				fields.mysqlUserRepository.EXPECT().SingleWithFilter(gomock.Any(),gomock.Any(),gomock.Any(),gomock.Any(),gomock.Any(),gomock.Any()).Return(nil)
				fields.mysqlUserRepository.EXPECT().UpdateSelectedField(gomock.Any(), []string{"premium_expires_at","updated_at"},gomock.Any(),args.userId).Return(nil)

				return fields
			},
			args: args{
				beegoCtx:   contextBeego,
				userId: 0,
				token:      "aaa",
				values: map[string]interface{}{
					"premium_expires_at" : time.Now().AddDate(0,1,0),
					"updated_at" : time.Now(),
				},
			},
		},
		{
			name:    "error context deadline exceeded SingleWithFilter",
			wantErr:  func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.EqualError(t, err, "context deadline exceeded")
			},
			fields: func(args *args,ctrl *gomock.Controller) fields {
				fields := toField(ctrl)
				fields.mysqlUserRepository.EXPECT().SingleWithFilter(gomock.Any(),gomock.Any(),gomock.Any(),gomock.Any(),gomock.Any(),gomock.Any()).Return(errors.New("context deadline exceeded"))
				fields.zapLogger.EXPECT().SetMessageLog(errors.New("context deadline exceeded"))
				return fields
			},
			args: args{
				beegoCtx:   contextBeego,
				userId: 0,
				token:      "aaa",
				values: map[string]interface{}{
					"premium_expires_at" : time.Now().AddDate(0,1,0),
					"updated_at" : time.Now(),
				},
			},
		},
		{
			name:    "error context deadline exceeded UpdateSelectedField",
			wantErr:  func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.EqualError(t, err, "context deadline exceeded")
			},
			fields: func(args *args,ctrl *gomock.Controller) fields {
				fields := toField(ctrl)
				fields.mysqlUserRepository.EXPECT().SingleWithFilter(gomock.Any(),gomock.Any(),gomock.Any(),gomock.Any(),gomock.Any(),gomock.Any()).Return(nil)
				fields.mysqlUserRepository.EXPECT().UpdateSelectedField(gomock.Any(), []string{"premium_expires_at","updated_at"},gomock.Any(),args.userId).Return(errors.New("context deadline exceeded"))
				fields.zapLogger.EXPECT().SetMessageLog(errors.New("context deadline exceeded"))
				return fields
			},
			args: args{
				beegoCtx:   contextBeego,
				userId: 0,
				token:      "aaa",
				values: map[string]interface{}{
					"premium_expires_at" : time.Now().AddDate(0,1,0),
					"updated_at" : time.Now(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func() {
			ctrl := gomock.NewController(t.T())
			defer ctrl.Finish()


			fields := tt.fields(&tt.args, ctrl)
			r := userUseCase{
				zapLogger            :                       fields.zapLogger,
				jwtAuth                  :                       fields.jwtAuth,
				expireToken             :                       fields.expireToken,
				contextTimeout           :                       fields.contextTimeout,
				mysqlUserRepository   :                       fields.mysqlUserRepository,
				mysqlProfileRepository :                       fields.mysqlProfileRepository,
			}
			err := r.PurchasePremiumUpdateStatus(tt.args.beegoCtx)
			if !tt.wantErr(t.T(), err, fmt.Sprintf("PurchasePremiumUpdateStatus(%v)", tt.args.beegoCtx)) {
				return
			}
		})
	}
}


func TestUserUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(UserUseCaseTestSuite))
}
