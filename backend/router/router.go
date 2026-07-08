package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"minecraft-manager/config"
	"minecraft-manager/handler"
	"minecraft-manager/middleware"
	"minecraft-manager/pkg/rcon"
	"minecraft-manager/service"
	"minecraft-manager/websocket"
)

func Setup(cfg *config.Config, db *gorm.DB, rconClient *rcon.RCONClient, hub *websocket.Hub) *gin.Engine {
	r := gin.Default()

	// Middleware
	r.Use(middleware.CORS())

	// Services
	authSvc := service.NewAuthService(db, cfg.JWTSecret)
	playerSvc := service.NewPlayerService(rconClient, db)
	serverSvc := service.NewServerService(rconClient)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	playerH := handler.NewPlayerHandler(playerSvc, hub)
	consoleH := handler.NewConsoleHandler(rconClient, db, hub)
	serverH := handler.NewServerHandler(serverSvc)

	// Public routes
	api := r.Group("/api")
	{
		api.POST("/login", authH.Login)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		// Auth
		protected.POST("/logout", authH.Logout)
		protected.GET("/me", authH.GetMe)

		// Players
		protected.GET("/players", playerH.GetPlayers)
		protected.POST("/players/kick", playerH.KickPlayer)
		protected.POST("/players/ban", playerH.BanPlayer)
		protected.POST("/players/op", playerH.OpPlayer)
		protected.POST("/players/deop", playerH.DeopPlayer)

		// Console
		protected.POST("/console/exec", consoleH.ExecCommand)

		// Server status
		protected.GET("/server/status", serverH.GetStatus)
	}

	// WebSocket
	r.GET("/ws", websocket.HandleWS(hub, cfg.JWTSecret))

	return r
}
