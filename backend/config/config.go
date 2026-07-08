package config

import "os"

type Config struct {
	ServerPort string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	RedisAddr  string
	RedisPass  string
	JWTSecret  string
	RCONHost   string
	RCONPort   int
	RCONPass   string
	LogPath    string
}

func Load() *Config {
	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "mcuser"),
		DBPassword: getEnv("DB_PASSWORD", "mcpass123"),
		DBName:     getEnv("DB_NAME", "minecraft_manager"),
		RedisAddr:  getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPass:  getEnv("REDIS_PASSWORD", ""),
		JWTSecret:  getEnv("JWT_SECRET", "minecraft-manager-secret-key-change-in-production"),
		RCONHost:   getEnv("RCON_HOST", "127.0.0.1"),
		RCONPort:   25575,
		RCONPass:   getEnv("RCON_PASSWORD", ""),
		LogPath:    getEnv("LOG_PATH", "/var/log/minecraft.log"),
	}
}

func (c *Config) DSN() string {
	return c.DBUser + ":" + c.DBPassword + "@tcp(" + c.DBHost + ":" + c.DBPort + ")/" + c.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
