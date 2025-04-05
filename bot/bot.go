package bot

import (
	"fmt"
	"log"

	"regexp"
	"strconv"
	"sync"
	"time"
	"urentBot/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
	StatScauts        = make(map[int64]WendayScaut)
	Scauts            = make(map[int64]Scaut)
	Subs              = make(map[string]int64)
	Dublers           = make(map[string]int)
	reportDubler      int
	IdealUnitsPerHour = float64(15.5)
)

func (b *Bot) Start() {
	log.Printf("Bot started: %s!", b.bot.Self.FirstName)

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
		if upd.Message.Text != "" && upd.Message.Photo == nil {
			b.DublersHandler(*upd.Message)
		}
	}
}

func (b *Bot) ResetReportRGL(timeReset time.Duration) {
	for {
		time.Sleep(b.cfg.TimeReset)
		b.muScauts.Lock()
		for key, scaut := range Scauts {
			if !scaut.TimeStart.IsZero() {
				if time.Until(scaut.TimeStart.Add(b.cfg.TimeReset)) < 0 {
					MsgForAdmin := tgbotapi.NewMessage(b.cfg.AdminChannel, "")
					MsgForAdmin.Text = b.GenerateReportRGL(scaut)

					b.AddStat(scaut, key)

					if Subs[scaut.UserName] != 0 {
						msgScaut := tgbotapi.NewMessage(Subs[scaut.UserName], "")
						msgScaut.Text = b.GenerateReportScaut(scaut)
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

func (b *Bot) DublersHandler(msg tgbotapi.Message) {
	reportDubler++
	AddToDublers(msg.Text)
	if reportDubler >= 3 {
		if ResultDubler := findDublicates(Dublers); ResultDubler != nil {
			message := tgbotapi.NewMessage(b.cfg.AdminChannel, "")
			for key, value := range ResultDubler {
				message.Text += fmt.Sprintf(" - %s: %d Раз(а)\n", key, value-1)
			}
			if message.Text != "" {
				message.Text = "⛔ Дубликаты найдены ⛔\n" + message.Text
				message.Text += "@" + msg.From.UserName
				b.bot.Send(message)
			}
		}
		reportDubler = 0
		Dublers = make(map[string]int)
		AddToDublers(msg.Text)

	}

}
func AddToDublers(msg string) {
	re, _ := regexp.Compile(`S\.\d{6}`)
	numbers := re.FindAllString(msg, -1)
	if numbers != nil && reportDubler < 3 {
		for _, n := range numbers {
			Dublers[n] += 1
		}
	}
}

func findDublicates(dublers map[string]int) map[string]int {
	dubler := make(map[string]int)
	for key, value := range dublers {
		if value > 1 {
			dubler[key] += value
		}
	}
	return dubler
}

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

func SerchMoved(msg string) int {
	re := regexp.MustCompile(`(?i)(Перемещения|Итого)[:\-\s]*(\d+)`)
	matches := re.FindStringSubmatch(msg)
	if len(matches) < 3 {
		return 0
	}
	result, _ := strconv.Atoi(matches[2])
	return result
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

func getDate() string {
	now := time.Now()
	return fmt.Sprintf("%02d.%02d", now.Day(), now.Month())
}

func (b *Bot) getTimeReport(start time.Time) string {
	future := start.Add(time.Duration(b.cfg.TimeLocation) * time.Hour)
	hour, min := future.Hour(), future.Minute()
	return fmt.Sprintf("%02d:%02d", hour, min)
}

func efficiencyPercent(hours, moves float64) float64 {
	actualUnitsPerHour := float64(moves) / float64(hours)
	return (actualUnitsPerHour / IdealUnitsPerHour) * 100
}
