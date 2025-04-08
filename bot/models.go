package bot

import (
	"time"
)

type WendayScaut struct {
	UserName       string // Тг ник
	SummerHour     int    // Количество часов всего
	SummerLateness int    // Количество опозданий
	SummerMuved    int    // Количество перемещений
}

type Scaut struct {
	Moved     int
	Images    int
	UserName  string
	Lateness  int
	TimeStart time.Time
	FirstTime time.Time
	ChatID    int64
}

type DublersTracker struct {
	data    map[string]int
	counter int
}
