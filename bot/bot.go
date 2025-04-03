package bot

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"urentBot/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TODO: Append Wenday Statistic by Scauts and Send For Chanel

type Bot struct {
	cfg *config.Config
	bot *tgbotapi.BotAPI
	*Mut
	wg *sync.WaitGroup
}

type Mut struct {
	muScauts sync.RWMutex
	muStats  sync.RWMutex
}

func NewBot(cfg config.Config) *Bot {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Fatal(err)
	}
	return &Bot{
		cfg: &cfg,
		bot: bot,
		Mut: &Mut{
			muScauts: sync.RWMutex{},
			muStats:  sync.RWMutex{},
		},
		wg: &sync.WaitGroup{},
	}
}

var (
	StatScauts = make(map[int64]WendayScaut)
	Scauts     = make(map[int64]Scaut)
	Subs       = make(map[string]int64)
)

func (b *Bot) Start() {

	go b.ResetReportRGL(b.cfg.TimeReset)
	updates := b.bot.GetUpdatesChan(tgbotapi.NewUpdate(0))
	for upd := range updates {
		if upd.Message == nil {
			continue
		}
		if upd.Message.IsCommand() {
			if upd.Message.Chat.Type != "private" {
				msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "")
				del := tgbotapi.NewDeleteMessage(msg.ChatID, upd.Message.MessageID)
				b.bot.Send(del)
				continue
			}
			b.CMDHanlder(*upd.Message)
			continue
		}
		b.MessageHandler(*upd.Message)
	}
}

func (b *Bot) ResetReportRGL(timeReset time.Duration) {
	for {
		time.Sleep(b.cfg.TimeReset)
		b.muScauts.Lock()
		for key, scaut := range Scauts {
			if !scaut.TimeStart.IsZero() {
				lastReport := scaut.TimeStart.Add(b.cfg.TimeReset)
				if lastReport.Sub(time.Now()) < 0 {
					MsgForAdmin := tgbotapi.NewMessage(b.cfg.AdminChannel, "")
					MsgForAdmin.Text = fmt.Sprintf("Смена завершенна Ботом!\n@%s:\nСмену завершил: %s.c%s-%s (%s)\nПереместил: %d\nНавёл порядок: %d\nОтчёты более 30 минут: %d\n\n", scaut.UserName, getDate(),

						b.getTimeReport(scaut.FirstTime), b.getTimeReport(scaut.TimeStart), scaut.TimeStart.Sub(scaut.FirstTime).String(),
						scaut.Moved, scaut.Images, scaut.Lateness,
					)
					b.AddStat(scaut, key)

					if Subs[scaut.UserName] != 0 {
						msgScaut := tgbotapi.NewMessage(Subs[scaut.UserName], "")
						sumHour := strings.Split(scaut.TimeStart.Sub(scaut.FirstTime).String(), "h")
						msgScaut.Text = fmt.Sprintf("Смену завершил %s.c %s-%s (%s Часов)\n🔁Перемещения: %d \n✅Навёл порядок: %d \nИтого: %d", getDate(),
							b.getTimeReport(scaut.FirstTime), b.getTimeReport(scaut.TimeStart), sumHour[0],
							scaut.Moved, scaut.Images, scaut.Lateness,
						)
						b.bot.Send(msgScaut)
					}
					Scauts[key] = Scaut{}
					b.bot.Send(MsgForAdmin)
				}
			}
		}
		b.muScauts.Unlock()

	}
}

// // OKEY
func (b *Bot) MessageHandler(msg tgbotapi.Message) {
	b.muScauts.Lock()
	defer b.muScauts.Unlock()
	scaut := Scauts[msg.From.ID]

	if scaut.FirstTime.IsZero() || scaut.TimeStart.IsZero() {
		scaut.FirstTime = msg.Time()
		scaut.TimeStart = msg.Time()
	}

	if msg.Time().Sub(scaut.TimeStart) > b.cfg.TimeOfset {
		scaut.Lateness += 1
		scaut.TimeStart = msg.Time()
	}
	scaut.TimeStart = msg.Time()

	scaut.UserName = msg.From.UserName
	if msg.Photo != nil {
		scaut.Images += 1
		Scauts[msg.From.ID] = scaut
		return
	}
	scaut.Moved += SerchMoved(msg.Text)
	Scauts[msg.From.ID] = scaut
}

func (b *Bot) AddStat(scaut Scaut, id int64) {
	b.muStats.Lock()
	defer b.muStats.Unlock()

	sumHour := int(scaut.TimeStart.Sub(scaut.FirstTime).Hours())
	wenday := StatScauts[id]
	wenday.UserName = scaut.UserName
	wenday.SummerMuved += scaut.Moved
	wenday.SummerLateness += scaut.Lateness
	wenday.SummerHour += sumHour
	StatScauts[id] = wenday

}

// Helper Func
func SerchMoved(msg string) int {
	re := regexp.MustCompile(`(?i)(Перемещения|Итого)[:\-\s]*(\d+)`)
	matches := re.FindStringSubmatch(msg)
	if len(matches) < 3 {
		return 0
	}
	result, _ := strconv.Atoi(matches[2])
	return result
}

func getDate() string {
	now := time.Now()
	return fmt.Sprintf("%02d.%02d", now.Day(), now.Month())
}

func (b *Bot) getTimeReport(start time.Time) string {
	future := start.Add(time.Duration(b.cfg.TimeLocation) * time.Hour)
	hour, min := future.Hour(), future.Minute()
	return fmt.Sprintf("%02d:%02d", hour, min)
}
