package config

import "os"

type Config struct {
	DSN       string
	RedisAddr string
	Addr      string
	JWTSecret string
}

func Load() *Config {
	return &Config{
		DSN:       getEnv("DB_DSN", "root:root@tcp(127.0.0.1:13306)/im?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr: getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		Addr:      getEnv("SERVER_ADDR", ":8080"),
		JWTSecret: getEnv("JWT_SECRET", "im-secret-key-1018"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
