package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/radyatamaa/dating-apps-api/internal/domain"
	"github.com/radyatamaa/dating-apps-api/pkg/helper"
	"github.com/radyatamaa/dating-apps-api/pkg/jwt"
	"github.com/radyatamaa/dating-apps-api/pkg/validator"

	"github.com/beego/beego/v2/client/cache"
	_ "github.com/beego/beego/v2/client/cache/redis"

	"github.com/radyatamaa/dating-apps-api/internal"
	"github.com/radyatamaa/dating-apps-api/pkg/database"

	beego "github.com/beego/beego/v2/server/web"
	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/beego/beego/v2/server/web/filter/cors"
	"github.com/beego/i18n"
	"github.com/radyatamaa/dating-apps-api/internal/middlewares"
	"github.com/radyatamaa/dating-apps-api/pkg/response"
	"github.com/radyatamaa/dating-apps-api/pkg/zaplogger"

	userHandler "github.com/radyatamaa/dating-apps-api/internal/user/delivery/http/v1"
	userUsecase "github.com/radyatamaa/dating-apps-api/internal/user/usecase"
	userRepository "github.com/radyatamaa/dating-apps-api/internal/user/repository"

	profileHandler "github.com/radyatamaa/dating-apps-api/internal/profile/delivery/http/v1"
	profileUsecase "github.com/radyatamaa/dating-apps-api/internal/profile/usecase"
	profileRepository "github.com/radyatamaa/dating-apps-api/internal/profile/repository"

	swipeHandler "github.com/radyatamaa/dating-apps-api/internal/swipe/delivery/http/v1"
	swipeUsecase "github.com/radyatamaa/dating-apps-api/internal/swipe/usecase"
	swipeRepository "github.com/radyatamaa/dating-apps-api/internal/swipe/repository"
)

// @title Api Gateway V1
// @version v1
// @contact.name radyatama
// @contact.email mohradyatama24@gmail.com
// @description api "API Gateway v1"
// @BasePath /api
// @query.collection.format multi
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	err := beego.LoadAppConfig("ini", "conf/app.ini")
	if err != nil {
		panic(err)
	}
	// token expired
	tokenExpired := beego.AppConfig.DefaultInt64("tokenExpired", 86400)
	// global execution timeout
	serverTimeout := beego.AppConfig.DefaultInt64("serverTimeout", 60)
	// global execution timeout
	requestTimeout := beego.AppConfig.DefaultInt("executionTimeout", 5)
	// global execution timeout to second
	timeoutContext := time.Duration(requestTimeout) * time.Second
	// web hook to slack error log
	slackWebHookUrl := beego.AppConfig.DefaultString("slackWebhookUrlLog", "")
	// app version
	appVersion := beego.AppConfig.DefaultString("version", "1")
	// log path
	logPath := beego.AppConfig.DefaultString("logPath", "./logs/api.log")
	// redis connection config
	redisConnectionConfig := beego.AppConfig.DefaultString("redisBeegoConConfig", `{"conn":"127.0.0.1:6379"}`)
	// jwt secret key
	jwtSecretKey := beego.AppConfig.DefaultString("jwtSecretKey", "secret")
	// init data
	initDataDummyProfileSeeder := beego.AppConfig.DefaultString("initDataDummyProfileSeeder", "true")

	// database initialization
	db := database.DB()

	// language
	lang := beego.AppConfig.DefaultString("lang", "en|id")
	languages := strings.Split(lang, "|")
	for _, value := range languages {
		if err := i18n.SetMessage(value, "./conf/"+value+".ini"); err != nil {
			panic("Failed to set message file for l10n")
		}
	}

	// beego config
	beego.BConfig.Log.AccessLogs = false
	beego.BConfig.Log.EnableStaticLogs = false
	beego.BConfig.Listen.ServerTimeOut = serverTimeout

	// zap logger
	zapLog := zaplogger.NewZapLogger(logPath, slackWebHookUrl)

	if beego.BConfig.RunMode == "dev" {
		// db auto migrate dev environment
		if err := db.AutoMigrate(
			&domain.User{},
			&domain.Profile{},
			&domain.Swipe{},
		); err != nil {
			panic(err)
		}

		// static files swagger
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	// init redis
	redisCache, err := cache.NewCache("redis", redisConnectionConfig)

	if err != nil {
		panic(err)
	}

	// config validator
	validator.Validate.SetDatabaseConnection(db)

	// jwt middleware
	auth, err := jwt.NewJwt(&jwt.Options{
		SignMethod:  jwt.HS256,
		SecretKey:   jwtSecretKey,
		Locations:   "header:Authorization",
		IdentityKey: "uid",
	})
	if err != nil {
		panic(err)
	}

	// set adapter redis for jwt middleware
	auth.SetAdapter(redisCache)
	if initDataDummyProfileSeeder == "true" {
		domain.SeederDataUserProfile(db)
	}
	if beego.BConfig.RunMode != "prod" {
		// static files swagger
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	beego.BConfig.RecoverFunc = func(context *beegoContext.Context, config *beego.Config) {
		if err := recover(); err != nil {
			var stack string

			hasIndent := beego.BConfig.RunMode != beego.PROD
			out := response.ApiResponse{
				Code:      response.ServerErrorCode,
				Message:   response.ErrorCodeText(response.ServerErrorCode, helper.GetLangVersion(context)),
				Data:      nil,
				Errors:    nil,
				RequestId: context.ResponseWriter.ResponseWriter.Header().Get("X-REQUEST-ID"),
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			}

			for i := 1; ; i++ {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				stack = stack + fmt.Sprintln(fmt.Sprintf("%s:%d", file, line))
			}

			context.Input.SetData("stackTrace", &zaplogger.ListErrors{
				Error: stack,
				Extra: err,
			})

			if context.Output.Status != 0 {
				context.ResponseWriter.WriteHeader(context.Output.Status)
			} else {
				context.ResponseWriter.WriteHeader(500)
			}
			context.Output.JSON(out, hasIndent, false)
		}
	}

	beego.BConfig.WebConfig.StaticDir["/external"] = "external"

	// middleware init
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowMethods:    []string{http.MethodGet, http.MethodPost},
		AllowAllOrigins: true,
	}))

	beego.InsertFilterChain("*", middlewares.RequestID())
	beego.InsertFilterChain("/api/*", middlewares.BodyDumpWithConfig(middlewares.NewAccessLogMiddleware(zapLog, appVersion).Logger()))
	beego.InsertFilterChain("/api/v1/*", middlewares.NewJwtMiddleware().JwtMiddleware(auth))
	// health check
	beego.Get("/health", func(ctx *beegoContext.Context) {
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(beego.M{"status": "alive"}, beego.BConfig.RunMode != "prod", false)
	})

	// default error handler
	beego.ErrorController(&response.ErrorController{})

	// init repository
	userMysqlRepo := userRepository.NewMysqlRepository(db,zapLog)
	profileMysqlRepo := profileRepository.NewMysqlRepository(db,zapLog)
	swipeMysqlRepo := swipeRepository.NewMysqlRepository(db,zapLog)

	// init usecase
	userUseCase := userUsecase.NewUserUseCase(timeoutContext,userMysqlRepo,profileMysqlRepo,auth,int(tokenExpired),zapLog)
	profileUseCase := profileUsecase.NewProfileUseCase(timeoutContext,profileMysqlRepo,swipeMysqlRepo,zapLog)
	swipeUseCase := swipeUsecase.NewSwipeUseCase(timeoutContext,swipeMysqlRepo,userMysqlRepo,zapLog)

	// init handler
	userHandler.NewUserHandler(userUseCase,zapLog)
	profileHandler.NewProfileHandler(profileUseCase,zapLog)
	swipeHandler.NewSwipeHandler(swipeUseCase,zapLog)

	beego.BeeApp.Server.RegisterOnShutdown(func() {
		if sqlDb, err := db.DB(); err != nil {
			log.Println("error database connection ...")
		} else {
			if err := sqlDb.Close(); err != nil {
				log.Println("failed close database")
			} else {
				log.Println("close database connection ...")
			}
		}
	})

	// default error handler
	beego.ErrorController(&internal.BaseController{})

	beego.Run()
}
