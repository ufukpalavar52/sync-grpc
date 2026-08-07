package config

import (
	"fmt"
	"os"
	"strconv"
)

var AppName = GetEnvString("APP_NAME", "imap-sync")
var AppEnv = GetEnvString("APP_ENV", "dev")
var AppPort = GetEnvInt("APP_PORT", 50051)

var DBHost = GetEnvString("DB_HOST", "localhost")
var DBUser = GetEnvString("DB_USER", "")
var DBPassword = GetEnvString("DB_PASSWORD", "")
var DBName = GetEnvString("DB_DATABASE", "")
var DBPort = GetEnvInt("DB_PORT", 5432)
var DBTimezone = GetEnvString("DB_TIMEZONE", "Europe/Istanbul")
var DBSslMode = GetEnvInt("DB_SSL_MODE", 0)
var DBPoolMaxIdleConn = GetEnvInt("DB_POOL_MAX_IDLE_CONN", 10)
var DBPoolMaxOpenConn = GetEnvInt("DB_POOL_MAX_OPEN_CONN", 100)

var EncryptionKeyVersion = GetEnvString("EncryptionKeyVersion", "v1")

var SyncLogPath = GetEnvString("SYNC_LOG_PATH", "")

var LogPath = GetEnvString("LOG_PATH", "logs/app.log")
var LogLevel = GetEnvString("LOG_LEVEL", "info")

func GetEnv(key string, defaultVal any) any {
	value := os.Getenv(key)
	if value == "" {
		return defaultVal
	}
	return value
}

func GetEnvString(key string, defaultVal string) string {
	return fmt.Sprintf("%v", GetEnv(key, defaultVal))
}

func GetEnvInt(key string, defaultVal int) int {
	value := GetEnv(key, defaultVal)
	v, ok := value.(int)
	if ok {
		return v
	}

	str, ok := value.(string)
	if !ok {
		return defaultVal
	}

	v, err := strconv.Atoi(str)
	if err != nil {
		return defaultVal
	}

	return v
}
