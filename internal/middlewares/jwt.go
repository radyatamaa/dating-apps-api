package middlewares

import (
	"net/http"
	"strings"

	"github.com/radyatamaa/dating-apps-api/pkg/helper"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/radyatamaa/dating-apps-api/pkg/jwt"
	"github.com/radyatamaa/dating-apps-api/pkg/response"
)

type JwtConfig struct {
	Skipper Skipper
	response.ApiResponse
}

func NewJwtMiddleware() *JwtConfig {
	return &JwtConfig{Skipper: func(ctx *context.Context) bool {
		if strings.EqualFold(ctx.Request.URL.Path, "/api/v1/user/login") {
			return true
		}
		if strings.EqualFold(ctx.Request.URL.Path, "/api/v1/user/register") {
			return true
		}
		return false
	}}
}

func (r *JwtConfig) JwtMiddleware(jwtAuth jwt.JWT) beego.FilterChain {
	return func(next beego.FilterFunc) beego.FilterFunc {
		return func(ctx *context.Context) {
			if r.Skipper(ctx) {
				next(ctx)
				return
			}
			if ctx.Request.Method == "OPTIONS" {
				next(ctx)
				return
			}

			if middlewareRequest, err := jwtAuth.Middleware(ctx.Request); err != nil {
				switch {
				case jwt.IsInvalidToken(err):
					r.ResponseError(ctx, http.StatusUnauthorized, response.InvalidTokenCodeError, response.ErrorCodeText(response.InvalidTokenCodeError, helper.GetLangVersion(ctx)), err)
					return
				case jwt.IsExpiredToken(err):
					r.ResponseError(ctx, http.StatusUnauthorized, response.ExpiredTokenCodeError, response.ErrorCodeText(response.ExpiredTokenCodeError, helper.GetLangVersion(ctx)), err)
					return
				case jwt.IsMissingToken(err):
					r.ResponseError(ctx, http.StatusUnauthorized, response.MissingTokenCodeError, response.ErrorCodeText(response.MissingTokenCodeError, helper.GetLangVersion(ctx)), err)
					return
				case jwt.IsAuthElsewhere(err):
					r.ResponseError(ctx, http.StatusUnauthorized, response.AuthElseWhereCodeError, response.ErrorCodeText(response.AuthElseWhereCodeError, helper.GetLangVersion(ctx)), err)
					return
				default:
					r.ResponseError(ctx, http.StatusUnauthorized, response.UnauthorizedCodeError, response.ErrorCodeText(response.UnauthorizedCodeError, helper.GetLangVersion(ctx)), err)
					return
				}
			} else {
				ctx.Request = middlewareRequest
				next(ctx)
			}
		}
	}
}
