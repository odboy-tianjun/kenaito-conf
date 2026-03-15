package main

import (
	"flag"
	"fmt"
	"kenaito-conf/server/internal/config"
	"kenaito-conf/server/internal/dao"
	"kenaito-conf/server/internal/handler"
	"kenaito-conf/server/internal/middleware"
	"kenaito-conf/server/internal/service"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	configPath := flag.String("config", "application.properties", "config file path")
	flag.Parse()

	// load config
	if err := config.Load(*configPath); err != nil {
		log.Fatalf("load config error: %v", err)
	}

	// init database
	if err := dao.Init(); err != nil {
		log.Fatalf("init database error: %v", err)
	}
	defer dao.Engine.Close()

	// init services
	authService := service.NewAuthService(config.Cfg.JWT.Secret, config.Cfg.JWT.ExpireHours)
	parseService := service.NewConfigParseService()
	permService := service.NewPermissionService()

	// init handlers
	authHandler := handler.NewAuthHandler(authService)
	indexHandler := handler.NewIndexHandler()
	configHandler := handler.NewConfigHandler(parseService, permService)
	configApiHandler := handler.NewConfigApiHandler(parseService)
	appHandler := handler.NewAppHandler()
	userHandler := handler.NewUserHandler(permService)
	groupHandler := handler.NewGroupHandler()
	profileHandler := handler.NewProfileHandler()

	// init gin
	r := gin.Default()
	r.Use(middleware.CORS())

	// public routes
	r.POST("/user/login", authHandler.Login)

	// admin routes (JWT auth required)
	admin := r.Group("/admin")
	admin.Use(middleware.JWTAuth(authService))
	{
		admin.GET("/menus", indexHandler.GetMenus)

		// config management
		admin.POST("/config/list", configHandler.List)
		admin.GET("/config/:id", configHandler.GetById)
		admin.POST("/config/add", configHandler.Add)
		admin.POST("/config/update", configHandler.Update)
		admin.POST("/config/delete/:id", configHandler.Delete)
		admin.GET("/config/configHistories/:id", configHandler.GetHistory)
		admin.POST("/config/rollback/:id", configHandler.Rollback)

		// app management
		admin.POST("/app/list", appHandler.List)
		admin.GET("/app/:id", appHandler.GetById)
		admin.POST("/app/add", appHandler.Add)
		admin.POST("/app/update", appHandler.Update)
		admin.POST("/app/delete/:id", appHandler.Delete)
		admin.POST("/app/switch/:id", appHandler.Switch)
		admin.POST("/app/generateToken/:id", appHandler.GenerateToken)
		admin.POST("/app/deleteToken/:id", appHandler.DeleteToken)
		admin.GET("/app/options", appHandler.Options)

		// user management
		admin.GET("/user/list", userHandler.List)
		admin.POST("/user/add", userHandler.Add)
		admin.POST("/user/update", userHandler.AdminUpdate)
		admin.POST("/user/delete/:id", userHandler.Delete)
		admin.POST("/user/resetPassword/:id", userHandler.ResetPassword)
		admin.POST("/user/switch/:id", userHandler.Switch)
		admin.GET("/user/options", userHandler.Options)

		// group management
		admin.GET("/group/list", groupHandler.List)
		admin.GET("/group/get/:id", groupHandler.GetById)
		admin.POST("/group/add", groupHandler.Add)
		admin.POST("/group/update", groupHandler.Update)
		admin.POST("/group/switch/:id", groupHandler.Switch)
		admin.GET("/group/options", groupHandler.Options)

		// profile management
		admin.GET("/profile/list", profileHandler.List)
		admin.GET("/profile/:id", profileHandler.GetById)
		admin.POST("/profile/add", profileHandler.Add)
		admin.POST("/profile/update", profileHandler.Update)
		admin.POST("/profile/delete/:id", profileHandler.Delete)
		admin.POST("/profile/switch/:id", profileHandler.Switch)
		admin.GET("/profile/options", profileHandler.Options)
	}

	// user info routes (JWT auth required)
	user := r.Group("/user")
	user.Use(middleware.JWTAuth(authService))
	{
		user.GET("/loginUserInfo", authHandler.GetLoginUserInfo)
		user.POST("/logout", authHandler.Logout)
	}

	// client API routes (no JWT, uses token or open access)
	api := r.Group("/api")
	{
		api.POST("/fetchAllConfigs", configApiHandler.FetchAllConfigs)
		api.POST("/fetchChangedConfigs", configApiHandler.FetchChangedConfigs)
		api.GET("/ping", configApiHandler.Ping)
	}

	// start server
	addr := fmt.Sprintf(":%d", config.Cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
