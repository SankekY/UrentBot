package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

var Scauts = make(map[int64]Scaut)
var MsgForAdmin = tgbotapi.NewMessage(-1002464733218, "")

func initConfig() *Config {
	return &Config{
		Token:     "6195617199:AAGCTj4THILrqOn6ctd1_2OcGrQZAB0HVtA",
		TimeOfset: time.Duration(35 * time.Minute),
		TimeReset: time.Duration(90 * time.Minute),
	}
}

func initBot(cfg *Config) *B {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = false
	return &B{
		bot: bot,
		cfg: cfg,
	}
}

func main() {
	cfg := initConfig()
	b := initBot(cfg)
	go b.ResetReportRGL(cfg.TimeReset)

	updates := b.bot.GetUpdatesChan(tgbotapi.NewUpdate(0))
	for upd := range updates {
		if upd.Message.Text == "" && upd.Message.Photo == nil {
			continue
		}
		msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "")
		if upd.Message.IsCommand() {
			if upd.Message.Chat.Type != "private" {
				del := tgbotapi.NewDeleteMessage(msg.ChatID, upd.Message.MessageID)
				b.bot.Send(del)
				continue
			}

			switch upd.Message.Command() {
			case "start":
				msg.Text = "/info - Ифнормация по перемещениям и фото !\n /report - Сгенерировать отчёт"
				b.bot.Send(msg)
				continue
			case "info":
				user := Scauts[upd.Message.From.ID]

				msg.Text = fmt.Sprintf("Смену завершил %s.c %s-%s (%s)\n🔁Перемещения: %d \n✅Навёл порядок: %d \nИтого: %d",
					getDate(), getTimeReport(user.FirstTime), getTimeReport(user.TimeStart.Add(-b.cfg.TimeOfset)),
					user.TimeStart.Add(-b.cfg.TimeOfset).Sub(user.FirstTime).String(),
					user.Moved, user.Images, user.Moved+user.Images,
				)
				b.bot.Send(msg)
				continue
			case "report":
				go b.ReportScaut(*upd.Message)
				continue
			case "rgl":
				for _, value := range Scauts {
					start := value.TimeStart.Add(-b.cfg.TimeOfset)
					lastReport := getTimeReport(start)
					firstReport := getTimeReport(value.FirstTime)
					msg.Text = fmt.Sprintf("@%s: \nПереместил: %d\nНавёл порядок: %d\nОтчёты более 30 минут: %d\nПервый отчёт: %s\nПолследний отёт: %s\n\n",
						value.UserName, value.Moved, value.Images, value.Lateness, firstReport, lastReport,
					)
				}
				b.bot.Send(msg)
				continue
			default:
				continue
			}
		}
		b.MessageHandler(*upd.Message)
	}
}

func (b *B) ReportScaut(Message tgbotapi.Message) {
	date := getDate()
	scaut := Scauts[Message.From.ID]
	start := scaut.TimeStart.Add(-b.cfg.TimeOfset)
	lastReport := getTimeReport(start)

	// RGL Messsage
	if scaut.UserName != "" {
		MsgForAdmin.Text = fmt.Sprintf("@%s:\nСмену завершил: %s.c%s-%s (%s)\nПереместил: %d\nНавёл порядок: %d\nОтчёты более 30 минут: %d\n\n",
			scaut.UserName, date, getTimeReport(scaut.FirstTime), lastReport, start.Sub(scaut.FirstTime).String(), scaut.Moved, scaut.Images, scaut.Lateness,
		)
		b.bot.Send(MsgForAdmin)
	}

	// SCAUT Message
	sumHour := strings.Split(start.Sub(scaut.FirstTime).String(), "h")
	msg := tgbotapi.NewMessage(Message.Chat.ID, "")
	msg.Text = fmt.Sprintf("Смену завершил %s.c %d:%d-%s (%s Часов)\n🔁Перемещения: %d \n✅Навёл порядок: %d \nИтого: %d", date,
		scaut.FirstTime.Hour(), scaut.FirstTime.Minute(), lastReport, sumHour[0], scaut.Moved, scaut.Images, scaut.Moved+scaut.Images,
	)
	Scauts[Message.From.ID] = Scaut{}
	b.bot.Send(msg)
}

func (b *B) MessageHandler(msg tgbotapi.Message) {
	scaut := Scauts[msg.From.ID]
	if scaut.FirstTime.IsZero() {
		scaut.FirstTime = msg.Time()
	}
	if scaut.TimeStart.Sub(msg.Time()) < 0 && !scaut.TimeStart.IsZero() {
		scaut.Lateness += 1
		scaut.TimeStart = msg.Time().Add(b.cfg.TimeOfset)
	}
	scaut.TimeStart = msg.Time().Add(b.cfg.TimeOfset)
	scaut.UserName = msg.From.UserName
	if msg.Photo != nil {
		scaut.Images += 1
		Scauts[msg.From.ID] = scaut
		return
	}
	scaut.Moved += SerchMoved(msg.Text)
	Scauts[msg.From.ID] = scaut

}

func (b *B) ResetReportRGL(timeReset time.Duration) {
	for {
		time.Sleep(2 * time.Hour)
		for key, scaut := range Scauts {
			if !scaut.TimeStart.IsZero() {
				lastReport := scaut.TimeStart.Add(-b.cfg.TimeOfset + b.cfg.TimeReset)
				if lastReport.Sub(time.Now()) < 0 {
					date := getDate()
					start := scaut.TimeStart.Add(-b.cfg.TimeOfset)
					MsgForAdmin.Text = fmt.Sprintf("Смена завершенна Ботом!\n@%s:\nСмену завершил: %s.c%s-%s (%s)\nПереместил: %d\nНавёл порядок: %d\nОтчёты более 30 минут: %d\n\n",
						scaut.UserName, date, getTimeReport(scaut.FirstTime), getTimeReport(scaut.TimeStart.Add(-b.cfg.TimeOfset)), start.Sub(scaut.FirstTime).String(), scaut.Moved, scaut.Images, scaut.Lateness,
					)
					Scauts[key] = Scaut{}
					b.bot.Send(MsgForAdmin)
				}
			}
		}

	}
}

func SerchMoved(msg string) int {
	arrText := strings.Split(msg, "\n")
	NewArr := strings.Split(arrText[len(arrText)-1], " ")

	result, err := strconv.Atoi(NewArr[1])
	if err != nil {
		return 0
	}
	return result
}

func getDate() string {
	_, month, day := time.Now().Date()
	date := fmt.Sprintf("%d.%d", day, month)
	if day < 10 && month < 10 {
		return fmt.Sprintf("0%d.0%d", day, month)
	} else if day < 10 && month > 10 {
		return fmt.Sprintf("0%d.%d", day, month)
	} else if day > 10 && month < 10 {
		return fmt.Sprintf("%d.0%d", day, month)
	}
	return date
}

func getTimeReport(start time.Time) string {
	if start.Hour() < 10 && start.Minute() < 10 {
		return fmt.Sprintf("0%d:0%d", start.Hour(), start.Minute())

	} else if start.Hour() < 10 && start.Minute() > 10 {
		return fmt.Sprintf("0%d:%d", start.Hour(), start.Minute())

	} else if start.Hour() > 10 && start.Minute() < 10 {
		return fmt.Sprintf("%d:0%d", start.Hour(), start.Minute())
	}
	return fmt.Sprintf("%d:%d", start.Hour(), start.Minute())
}
