package main

import (
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type WendayScaut struct {
	UserName       string // Тг ник
	SummerHour     int    // Количество часов всего
	SummerLateness int    // Количество опозданий
	SummerMuved    int    // Количество перемещений
}

type Config struct {
	Token     string
	TimeOfset time.Duration
	TimeReset time.Duration
}

type B struct {
	cfg *Config
	bot *tgbotapi.BotAPI
}

type Scaut struct {
	Moved     int
	Images    int
	UserName  string
	Lateness  int
	TimeStart time.Time
	FirstTime time.Time
}
