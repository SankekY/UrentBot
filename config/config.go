package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Token        string
	AdminChannel int64
	Admins       map[int64]bool
	TimeOfset    time.Duration
	TimeReset    time.Duration
	TimeLocation time.Duration
}

func InitConfig() *Config {
	admins := strings.Split(os.Getenv("admin_array"), ",")
	AdminsMap := make(map[int64]bool)
	for _, admin := range admins {
		value, err := strconv.Atoi(admin)
		if err != nil {
			log.Println("Error load admin := ", admin, err)

		}
		AdminsMap[int64(value)] = true
	}
	return &Config{
		Token:        os.Getenv("bot_token"),
		AdminChannel: int64(getEnvInt("admin_channel_id", 0)),
		Admins:       AdminsMap,
		TimeOfset:    time.Duration(getEnvInt("", 35)) * time.Minute,
		TimeReset:    time.Duration(getEnvInt("", 90)) * time.Minute,
		TimeLocation: time.Duration(getEnvInt("", 0)),
	}
}

func getEnvInt(key string, defaultVal int) int {
	if value, ok := os.LookupEnv(key); ok {
		if result, err := strconv.Atoi(value); err == nil {
			return result
		}
		log.Printf("\nError %s not loaded !", key)
		return defaultVal
	}
	return defaultVal
}
