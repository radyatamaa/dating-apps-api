package v1

import (
	"context"
	"errors"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/radyatamaa/dating-apps-api/internal"
	"github.com/radyatamaa/dating-apps-api/internal/domain"
	"github.com/radyatamaa/dating-apps-api/internal/swipe"
	"github.com/radyatamaa/dating-apps-api/pkg/response"
	"github.com/radyatamaa/dating-apps-api/pkg/validator"
	"github.com/radyatamaa/dating-apps-api/pkg/zaplogger"
	"gorm.io/gorm"
	"net/http"
)

type SwipeHandler struct {
	ZapLogger zaplogger.Logger
	internal.BaseController
	response.ApiResponse
	Usecase swipe.UseCase
}

func NewSwipeHandler(useCase swipe.UseCase, zapLogger zaplogger.Logger) {
	pHandler := &SwipeHandler{
		ZapLogger: zapLogger,
		Usecase:   useCase,
	}
	beego.Router("/api/v1/swipe/profile", pHandler, "post:SwipeProfile")
}

func (h *SwipeHandler) Prepare() {
	// check user access when needed
	h.SetLangVersion()
}

// SwipeProfile
// @Title SwipeProfile
// @Tags Swipe
// @Summary SwipeProfile
// @Produce json
// @Security ApiKeyAuth
// @Param Accept-Language header string false "lang"
// @Success 200 {object} swagger.BaseResponse{errors=[]object,data=object}
// @Failure 400 {object} swagger.BadRequestErrorValidationResponse{errors=[]swagger.ValidationErrors,data=object}
// @Failure 408 {object} swagger.RequestTimeoutResponse{errors=[]object,data=object}
// @Failure 500 {object} swagger.InternalServerErrorResponse{errors=[]object,data=object}
// @Param body body domain.SwipeProfileRequest true "request payload"
// @Router /v1/swipe/profile [post]
func (h *SwipeHandler) SwipeProfile() {
	var request domain.SwipeProfileRequest

	if err := h.BindJSON(&request); err != nil {
		h.Ctx.Input.SetData("stackTrace", h.ZapLogger.SetMessageLog(err))
		h.ResponseError(h.Ctx, http.StatusBadRequest, response.ApiValidationCodeError, response.ErrorCodeText(response.ApiValidationCodeError, h.Locale.Lang), err)
		return
	}
	if err := validator.Validate.ValidateStruct(&request); err != nil {
		h.Ctx.Input.SetData("stackTrace", h.ZapLogger.SetMessageLog(err))
		h.ResponseError(h.Ctx, http.StatusBadRequest, response.ApiValidationCodeError, response.ErrorCodeText(response.ApiValidationCodeError, h.Locale.Lang), err)
		return
	}

	err := h.Usecase.SwipeProfile(h.Ctx, request)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			h.ResponseError(h.Ctx, http.StatusRequestTimeout, response.RequestTimeoutCodeError, response.ErrorCodeText(response.RequestTimeoutCodeError, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.DataNotFoundCodeError, response.ErrorCodeText(response.DataNotFoundCodeError, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, response.ErrLimitSwipeOrLike) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.LimitSwipeOrLikeErrorCode, response.ErrorCodeText(response.LimitSwipeOrLikeErrorCode, h.Locale.Lang), err)
			return
		}
		h.ResponseError(h.Ctx, http.StatusInternalServerError, response.ServerErrorCode, response.ErrorCodeText(response.ServerErrorCode, h.Locale.Lang), err)
		return
	}
	h.Ok(h.Ctx, h.Tr("message.success"), nil)
	return
}
