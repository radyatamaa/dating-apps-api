package v1

import (
	"context"
	"errors"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/radyatamaa/dating-apps-api/internal"
	"github.com/radyatamaa/dating-apps-api/internal/domain"
	"github.com/radyatamaa/dating-apps-api/internal/profile"
	"github.com/radyatamaa/dating-apps-api/pkg/database/paginator"
	"github.com/radyatamaa/dating-apps-api/pkg/response"
	"github.com/radyatamaa/dating-apps-api/pkg/validator"
	"github.com/radyatamaa/dating-apps-api/pkg/zaplogger"
	"net/http"
)

type ProfileHandler struct {
	ZapLogger zaplogger.Logger
	internal.BaseController
	response.ApiResponse
	Usecase profile.UseCase
}

func NewProfileHandler(useCase profile.UseCase, zapLogger zaplogger.Logger) {
	pHandler := &ProfileHandler{
		ZapLogger: zapLogger,
		Usecase:   useCase,
	}
	beego.Router("/api/v1/profile", pHandler, "get:GetProfiles")
	beego.Router("/api/v1/profile/location", pHandler, "put:UpdateLiveLocationProfiles")
}

func (h *ProfileHandler) Prepare() {
	// check user access when needed
	h.SetLangVersion()
}

// @Title GetProfiles
// @Tags Profile
// @Summary GetProfiles
// @Produce json
// @Security ApiKeyAuth
// @Param Accept-Language header string false "lang"
// @Success 200 {object} swagger.BaseResponse{errors=[]object,data=domain.GetProfilesResponsePaginationResponse}
// @Failure 400 {object} swagger.BadRequestErrorValidationResponse{errors=[]swagger.ValidationErrors,data=object}
// @Failure 408 {object} swagger.RequestTimeoutResponse{errors=[]object,data=object}
// @Failure 500 {object} swagger.InternalServerErrorResponse{errors=[]object,data=object}
// @Param pageSize query int false "page size"
// @Param page query int false "page"
// @Param latitude query float64 false "current latitude"
// @Param longitude query float64 false "current longitude"
// @Router /v1/profile [get]
func (h *ProfileHandler) GetProfiles() {
	pageSize, page, err := paginator.PaginationQueryParamValidation(h.Ctx.Input.Query("pageSize"), h.Ctx.Input.Query("page"))
	if err != nil {
		h.ResponseError(h.Ctx, http.StatusBadRequest, response.QueryParamInvalidCode, response.ErrorCodeText(response.QueryParamInvalidCode, h.Locale.Lang), err)
		return
	}
	limit, page, offset := paginator.Pagination(page, pageSize)

	result, err := h.Usecase.GetProfiles(h.Ctx, page, limit, offset,h.Ctx.Input.Query("latitude"),h.Ctx.Input.Query("longitude"))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			h.ResponseError(h.Ctx, http.StatusRequestTimeout, response.RequestTimeoutCodeError, response.ErrorCodeText(response.RequestTimeoutCodeError, h.Locale.Lang), err)
			return
		}
		h.ResponseError(h.Ctx, http.StatusInternalServerError, response.ServerErrorCode, response.ErrorCodeText(response.ServerErrorCode, h.Locale.Lang), err)
		return
	}
	h.Ok(h.Ctx, h.Tr("message.success"), result)
	return
}

// UpdateLiveLocationProfiles
// @Title UpdateLiveLocationProfiles
// @Tags Profile
// @Summary UpdateLiveLocationProfiles
// @Produce json
// @Security ApiKeyAuth
// @Param Accept-Language header string false "lang"
// @Success 200 {object} swagger.BaseResponse{errors=[]object,data=object}
// @Failure 400 {object} swagger.BadRequestErrorValidationResponse{errors=[]swagger.ValidationErrors,data=object}
// @Failure 408 {object} swagger.RequestTimeoutResponse{errors=[]object,data=object}
// @Failure 500 {object} swagger.InternalServerErrorResponse{errors=[]object,data=object}
// @Param body body domain.UpdateLiveLocationProfilesRequest true "request payload"
// @Router /v1/profile/location [put]
func (h *ProfileHandler) UpdateLiveLocationProfiles() {
	var request domain.UpdateLiveLocationProfilesRequest

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

	err := h.Usecase.UpdateLiveLocationProfiles(h.Ctx, request)
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
