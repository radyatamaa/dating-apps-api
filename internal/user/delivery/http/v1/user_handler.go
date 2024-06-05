package v1

import (
	"context"
	"errors"
	"github.com/radyatamaa/dating-apps-api/internal/user"
	"github.com/radyatamaa/dating-apps-api/pkg/helper"
	"net/http"

	"github.com/radyatamaa/dating-apps-api/pkg/validator"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/radyatamaa/dating-apps-api/internal"
	"github.com/radyatamaa/dating-apps-api/internal/domain"
	"github.com/radyatamaa/dating-apps-api/pkg/response"
	"github.com/radyatamaa/dating-apps-api/pkg/zaplogger"
	"gorm.io/gorm"
)

type UserHandler struct {
	ZapLogger zaplogger.Logger
	internal.BaseController
	response.ApiResponse
	Usecase user.UseCase
}

func NewUserHandler(useCase user.UseCase, zapLogger zaplogger.Logger) {
	pHandler := &UserHandler{
		ZapLogger: zapLogger,
		Usecase:   useCase,
	}
	beego.Router("/api/v1/user/login", pHandler, "post:Login")
	beego.Router("/api/v1/user/register", pHandler, "post:Register")
	beego.Router("/api/v1/user/purchase-premium", pHandler, "post:PurchasePremiumUpdateStatus")
}

func (h *UserHandler) Prepare() {
	// check user access when needed
	h.SetLangVersion()
}

// Login
// @Title Login
// @Tags User
// @Summary Login
// @Produce json
// @Param Accept-Language header string false "lang"
// @Success 200 {object} swagger.BaseResponse{errors=[]object,data=domain.LoginResponse}
// @Failure 400 {object} swagger.BadRequestErrorValidationResponse{errors=[]swagger.ValidationErrors,data=object}
// @Failure 408 {object} swagger.RequestTimeoutResponse{errors=[]object,data=object}
// @Failure 500 {object} swagger.InternalServerErrorResponse{errors=[]object,data=object}
// @Param body body domain.LoginRequest true "request payload"
// @Router /v1/user/login [post]
func (h *UserHandler) Login() {
	var request domain.LoginRequest

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

	result, err := h.Usecase.Login(h.Ctx, request)
	if err != nil {
		if errors.Is(err, response.ErrInvalidEmailPassword) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.InvalidEmailPasswordErrorCode, response.ErrorCodeText(response.InvalidEmailPasswordErrorCode, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			h.ResponseError(h.Ctx, http.StatusRequestTimeout, response.RequestTimeoutCodeError, response.ErrorCodeText(response.RequestTimeoutCodeError, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.DataNotFoundCodeError, response.ErrorCodeText(response.DataNotFoundCodeError, h.Locale.Lang), err)
			return
		}
		h.ResponseError(h.Ctx, http.StatusInternalServerError, response.ServerErrorCode, response.ErrorCodeText(response.ServerErrorCode, h.Locale.Lang), err)
		return
	}
	h.Ok(h.Ctx, h.Tr("message.success"), result)
	return
}


// Register
// @Title Register
// @Tags User
// @Summary Register
// @Produce json
// @Param Accept-Language header string false "lang"
// @Success 200 {object} swagger.BaseResponse{errors=[]object,data=object}
// @Failure 400 {object} swagger.BadRequestErrorValidationResponse{errors=[]swagger.ValidationErrors,data=object}
// @Failure 408 {object} swagger.RequestTimeoutResponse{errors=[]object,data=object}
// @Failure 500 {object} swagger.InternalServerErrorResponse{errors=[]object,data=object}
// @Param        photo   formData  file    true  "file"
// @Param        name    formData  string  true  "name"
// @Param        age    formData  int  true  "age"
// @Param        bio    formData  string  true  "bio"
// @Param        email    formData  string  true  "email"
// @Param        password    formData  string  true  "password"
// @Router /v1/user/register [post]
func (h *UserHandler) Register() {
	var request domain.RegisterRequest
	if err := h.BindForm(&request); err != nil {
		h.Ctx.Input.SetData("stackTrace", h.ZapLogger.SetMessageLog(err))
		h.ResponseError(h.Ctx, http.StatusBadRequest, response.ApiValidationCodeError, response.ErrorCodeText(response.ApiValidationCodeError, h.Locale.Lang), err)
		return
	}

	file, fileHeader, err := h.GetFile("photo")
	if err != nil {
		h.Ctx.Input.SetData("stackTrace", h.ZapLogger.SetMessageLog(err))
		h.ResponseError(h.Ctx, http.StatusBadRequest, response.ApiValidationCodeError, response.ErrorCodeText(response.ApiValidationCodeError, h.Locale.Lang), err)
		return
	}

	if err := helper.ValidateFile(fileHeader); err != nil {
		if errors.Is(err, helper.ErrInvalidFormatJpeg) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.InvalidFormatJpegErrorCode, response.ErrorCodeText(response.InvalidFormatJpegErrorCode, h.Locale.Lang), err)
			return
		}
		h.Ctx.Input.SetData("stackTrace", h.ZapLogger.SetMessageLog(err))
		h.ResponseError(h.Ctx, http.StatusBadRequest, response.ApiValidationCodeError, response.ErrorCodeText(response.ApiValidationCodeError, h.Locale.Lang), err)
		return
	}

	photoUrl,err := helper.UploadFileJpeg(h.Ctx,file)
	if err != nil {
		h.Ctx.Input.SetData("stackTrace", h.ZapLogger.SetMessageLog(err))
		h.ResponseError(h.Ctx, http.StatusBadRequest, response.ApiValidationCodeError, response.ErrorCodeText(response.ApiValidationCodeError, h.Locale.Lang), err)
		return
	}
	request.Photo = photoUrl

	if err := validator.Validate.ValidateStruct(&request); err != nil {
		h.Ctx.Input.SetData("stackTrace", h.ZapLogger.SetMessageLog(err))
		h.ResponseError(h.Ctx, http.StatusBadRequest, response.ApiValidationCodeError, response.ErrorCodeText(response.ApiValidationCodeError, h.Locale.Lang), err)
		return
	}

	err = h.Usecase.Register(h.Ctx, request)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			h.ResponseError(h.Ctx, http.StatusRequestTimeout, response.RequestTimeoutCodeError, response.ErrorCodeText(response.RequestTimeoutCodeError, h.Locale.Lang), err)
			return
		}
		h.ResponseError(h.Ctx, http.StatusInternalServerError, response.ServerErrorCode, response.ErrorCodeText(response.ServerErrorCode, h.Locale.Lang), err)
		return
	}
	h.Ok(h.Ctx, h.Tr("message.success"), nil)
	return
}


// PurchasePremiumUpdateStatus
// @Title PurchasePremiumUpdateStatus
// @Tags User
// @Summary PurchasePremiumUpdateStatus
// @Produce json
// @Security ApiKeyAuth
// @Param Accept-Language header string false "lang"
// @Success 200 {object} swagger.BaseResponse{errors=[]object,data=object}
// @Failure 400 {object} swagger.BadRequestErrorValidationResponse{errors=[]swagger.ValidationErrors,data=object}
// @Failure 408 {object} swagger.RequestTimeoutResponse{errors=[]object,data=object}
// @Failure 500 {object} swagger.InternalServerErrorResponse{errors=[]object,data=object}
// @Router /v1/user/purchase-premium [post]
func (h *UserHandler) PurchasePremiumUpdateStatus() {
	err := h.Usecase.PurchasePremiumUpdateStatus(h.Ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			h.ResponseError(h.Ctx, http.StatusRequestTimeout, response.RequestTimeoutCodeError, response.ErrorCodeText(response.RequestTimeoutCodeError, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.DataNotFoundCodeError, response.ErrorCodeText(response.DataNotFoundCodeError, h.Locale.Lang), err)
			return
		}
		h.ResponseError(h.Ctx, http.StatusInternalServerError, response.ServerErrorCode, response.ErrorCodeText(response.ServerErrorCode, h.Locale.Lang), err)
		return
	}
	h.Ok(h.Ctx, h.Tr("message.success"), nil)
	return
}