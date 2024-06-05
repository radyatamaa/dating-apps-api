package v1

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/radyatamaa/dating-apps-api/internal"
	"github.com/radyatamaa/dating-apps-api/internal/domain/mocks"
	"github.com/radyatamaa/dating-apps-api/pkg/helper"
	"github.com/radyatamaa/dating-apps-api/pkg/response"
	"github.com/radyatamaa/dating-apps-api/pkg/zaplogger"
	mockZaplogger "github.com/radyatamaa/dating-apps-api/pkg/zaplogger/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type UserHandlerTestSuite struct {
	suite.Suite
}

func (t *UserHandlerTestSuite) SetupSuite() {
}

type fields struct {
	ZapLogger zaplogger.Logger
	BaseController internal.BaseController
	ApiResponse response.ApiResponse
	Usecase *mocks.MockUserUseCase
}

func toField(ctrl *gomock.Controller) fields {
	return fields{
		ZapLogger:               mockZaplogger.NewMockLogger(ctrl),
		BaseController:          internal.BaseController{},
		ApiResponse:             response.ApiResponse{},
		Usecase: 				 mocks.NewMockUserUseCase(ctrl),
	}
}

func(t *UserHandlerTestSuite) TestUserHandler_PurchasePremiumUpdateStatus() {
	tests := []struct {
		name       string
		fields     func(ctrl *gomock.Controller) (f fields, r *http.Request, w *httptest.ResponseRecorder)
		statusCode int
	}{
		{
			name: "success",
			fields: func(ctrl *gomock.Controller) (f fields, r *http.Request, w *httptest.ResponseRecorder) {
				f = toField(ctrl)
				r = httptest.NewRequest(http.MethodPost, "/api/v1/user/purchase-premium", strings.NewReader("")).WithContext(context.TODO())
				w = httptest.NewRecorder()

				f.Usecase.EXPECT().PurchasePremiumUpdateStatus(gomock.Any()).Return(nil)

				return
			},
			statusCode: http.StatusOK,
		},
		{
			name: "error context deadline exceeded",
			fields: func(ctrl *gomock.Controller) (f fields, r *http.Request, w *httptest.ResponseRecorder) {
				f = toField(ctrl)
				r = httptest.NewRequest(http.MethodPost, "/api/v1/user/purchase-premium", strings.NewReader("")).WithContext(context.TODO())
				w = httptest.NewRecorder()

				f.Usecase.EXPECT().PurchasePremiumUpdateStatus(gomock.Any()).Return(context.DeadlineExceeded)

				return
			},
			statusCode: http.StatusRequestTimeout,
		},
		{
			name: "error ErrRecordNotFound",
			fields: func(ctrl *gomock.Controller) (f fields, r *http.Request, w *httptest.ResponseRecorder) {
				f = toField(ctrl)
				r = httptest.NewRequest(http.MethodPost, "/api/v1/user/purchase-premium", strings.NewReader("")).WithContext(context.TODO())
				w = httptest.NewRecorder()

				f.Usecase.EXPECT().PurchasePremiumUpdateStatus(gomock.Any()).Return(gorm.ErrRecordNotFound)

				return
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "error internal server",
			fields: func(ctrl *gomock.Controller) (f fields, r *http.Request, w *httptest.ResponseRecorder) {
				f = toField(ctrl)
				r = httptest.NewRequest(http.MethodPost, "/api/v1/user/purchase-premium", strings.NewReader("")).WithContext(context.TODO())
				w = httptest.NewRecorder()

				f.Usecase.EXPECT().PurchasePremiumUpdateStatus(gomock.Any()).Return(errors.New("error server"))

				return
			},
			statusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func() {
			ctrl := gomock.NewController(t.T())
			defer ctrl.Finish()
			f, r, w := tt.fields(ctrl)
			h := &UserHandler{
				ZapLogger:               f.ZapLogger,
				BaseController:          f.BaseController,
				ApiResponse:             f.ApiResponse,
				Usecase:   f.Usecase,
			}

			helper.PrepareHandler(&h.Controller, r, w)
			h.PurchasePremiumUpdateStatus()

			assert.Equal(t.T(), tt.statusCode, w.Code)
		})
	}
}

func TestUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}
