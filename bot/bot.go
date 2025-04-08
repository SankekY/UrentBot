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
	DublersTracker
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
			if upd.Message.Chat.Type != "private" && upd.Message.Chat.ID != b.cfg.AdminChannel {
				msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "")
				del := tgbotapi.NewDeleteMessage(msg.ChatID, upd.Message.MessageID)
				b.bot.Send(del)
				continue
			}
			b.CMDHanlder(*upd.Message)
			continue
		}
		if upd.Message.Chat.ID == b.cfg.ScautChannel {
			b.MessageHandler(*upd.Message)
			if upd.Message.Photo == nil {
				b.DoublersHandler(*upd.Message)
			}
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

func (b *Bot) DoublersHandler(msg tgbotapi.Message) {

	b.addToDoublers(msg.Text)
	if b.DublersTracker.counter >= 4 {
		if duplicates := b.findDuplicates(); len(duplicates) > 0 {
			message := tgbotapi.NewMessage(b.cfg.AdminChannel, "")
			for key, value := range duplicates {
				if value > 0 {
					message.Text += fmt.Sprintf("- <code>%s</code> : (%d)Раз \n", key, value)
				}
			}
			if message.Text != "" {
				message.ParseMode = "html"
				message.Text = "⛔ Дубликаты найдены ⛔\n" + message.Text + fmt.Sprintf("@%s", msg.From.UserName)
				b.bot.Send(message)
			}
		}
		b.DublersTracker.counter = 0
		b.DublersTracker.data = make(map[string]int)
		b.addToDoublers(msg.Text)
	}
}

func (b *Bot) addToDoublers(text string) {
	re := regexp.MustCompile(`S.\d{6}`) // Пример: S.123456
	numbers := re.FindAllString(text, -1)
	if numbers != nil {
		b.DublersTracker.counter++
		for _, n := range numbers {
			b.DublersTracker.data[n]++
		}
	}
}

func (b *Bot) findDuplicates() map[string]int {
	duplicates := make(map[string]int)
	for key, value := range b.DublersTracker.data {
		if value > 1 {
			duplicates[key] = value - 1 // Количество повторений
		}
	}
	return duplicates
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
