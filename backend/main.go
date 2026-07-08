package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"minecraft-manager/config"
	"minecraft-manager/model"
	"minecraft-manager/pkg/logger"
	"minecraft-manager/pkg/rcon"
	"minecraft-manager/pkg/redis"
	"minecraft-manager/router"
	"minecraft-manager/seed"
	"minecraft-manager/websocket"
)

func main() {
	// Initialize logger
	logger.Init()
	logger.Info.Println("Starting Minecraft Manager...")

	// Load config
	cfg := config.Load()

	// Connect to MySQL
	logger.Info.Println("Connecting to MySQL...")
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		logger.Error.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&model.User{}, &model.CommandLog{}, &model.BanRecord{}); err != nil {
		logger.Error.Fatalf("Failed to migrate database: %v", err)
	}
	logger.Info.Println("Database migration complete")

	if err := seed.DefaultAdmin(db); err != nil {
		logger.Error.Fatalf("Failed to seed default admin: %v", err)
	}
	logger.Info.Println("Default admin user ready")

	// Connect to Redis
	logger.Info.Println("Connecting to Redis...")
	if err := redis.Init(cfg.RedisAddr, cfg.RedisPass); err != nil {
		logger.Warn.Printf("Redis connection failed (non-fatal): %v", err)
	} else {
		logger.Info.Println("Redis connected")
		defer redis.Close()
	}

	// Initialize RCON client
	logger.Info.Println("Setting up RCON client...")
	rconClient := rcon.New(cfg.RCONHost, cfg.RCONPort, cfg.RCONPass)
	if err := rconClient.Connect(); err != nil {
		logger.Warn.Printf("RCON connection failed (server may be offline): %v", err)
	} else {
		logger.Info.Println("RCON connected")
	}
	defer rconClient.Close()

	// Initialize WebSocket hub
	hub := websocket.NewHub()

	// Start log watcher
	logWatcher := websocket.NewLogWatcher(hub, cfg.LogPath)
	logWatcher.Start()
	defer logWatcher.Stop()

	// Setup router
	r := router.Setup(cfg, db, rconClient, hub)

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logger.Info.Println("Shutting down...")
		rconClient.Close()
		redis.Close()
		os.Exit(0)
	}()

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	logger.Info.Printf("Server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		logger.Error.Fatalf("Server failed: %v", err)
	}
}
