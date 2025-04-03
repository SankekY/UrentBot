package bot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var AwaitMessage = map[string]string{
	"start_admin": `
		🔻 ___ RGL Mode ___ 🔻

		📊 Доступные команды:

		🛠 /rgl — Общая информация по активным Скаутам
		📈 /stats — Полная статистика:
		├ 🔄 Перемещения
		├ 📸 Фотографирования  
		├ 🕐 Часы работы
		├ 💥 Проёбы 
		└ (с последнего сброса)

		♻️ /restats — Полный сброс статистики
		🔺 ____________________ 🔺
	`,
	"start_scaut": `
	┏━━━━━━━━━━ *Scaut Mode* ━━━━━━━━━━┓
	│                                   │
	│ 📌 */info*   → Текущая статистика │
	│    (без сброса смены)             │
	│                                   │
	│ 📝 */report* → Завершить смену и  │
	│    сгенерировать отчёт            │
	│                                   │
	│ 🔔 */sub*    → Авто-напоминание   │
	│    (пришлёт отчёт в ЛС, если      │
	│    не завершить смену за 1.5 часа)│
	│                                   │
	┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
	`,
}

func (b *Bot) CMDHanlder(msg tgbotapi.Message) {
	b.wg.Add(1)
	switch msg.Command() {
	case "start":
		go b.CMDStart(msg)

	case "info":
		go b.CMDInfo(msg)

	case "report":
		go b.CMDReport(msg)
	case "sub":
		b.CMDSubs(msg)
	default:
		if b.cfg.Admins[msg.From.ID] {
			switch msg.Command() {
			case "rgl":
				go b.CMDRgl(msg)
			case "stats":
				go b.CMDStats(msg)

			case "restats":
				go b.CMDRestats(msg)

			default:
				b.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Не понятная комманда напишите \n/start -Для просмтора комманд"))
			}

		}

	}
}

func (b *Bot) CMDStart(msg tgbotapi.Message) {
	defer b.wg.Done()

	if !b.cfg.Admins[msg.From.ID] {
		b.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, AwaitMessage["start_scaut"]))
	}
	b.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, AwaitMessage["start_admin"]))
}

func (b *Bot) CMDSubs(msg tgbotapi.Message) {
	subber, ok := Subs[msg.From.UserName]
	if ok {
		if subber == 0 {
			Subs[msg.From.UserName] = msg.Chat.ID
			b.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Вы подписанны на отчёты :)"))
			return
		}
		Subs[msg.From.UserName] = 0
		b.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Вы отписаны на отчёты :)"))
		return
	}
	Subs[msg.From.UserName] = msg.Chat.ID
	b.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Вы потписаны на отчёты :)"))

}

func (b *Bot) CMDRgl(msg tgbotapi.Message) {
	b.muScauts.Lock()
	defer b.wg.Done()
	defer b.muScauts.Unlock()

	message := tgbotapi.NewMessage(msg.Chat.ID, "📊 Текущая активность скаутов\n\n")
	for _, value := range Scauts {
		if value.FirstTime.IsZero() {
			continue
		}
		start := value.TimeStart
		lastReport := b.getTimeReport(start)
		firstReport := b.getTimeReport(value.FirstTime)
		message.Text += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
		message.Text += fmt.Sprintf("👤 @%s:\n🔄 Переместил: %d\n📸 Навёл порядок: %d\n⏱ Отчёты более 30 минут: %d\n⏳ Время работы: %s - %s \n\n",
			value.UserName, value.Moved, value.Images, value.Lateness, firstReport, lastReport,
		)
	}
	b.bot.Send(message)
}

func (b *Bot) CMDReport(msg tgbotapi.Message) {
	b.muScauts.Lock()
	defer b.wg.Done()
	defer b.muScauts.Unlock()

	scaut := Scauts[msg.From.ID]

	// RGL Messsage
	if scaut.UserName != "" {
		MsgForAdmin := tgbotapi.NewMessage(b.cfg.AdminChannel, "")

		MsgForAdmin.Text = fmt.Sprintf("@%s:\nСмену завершил: %s.c%s-%s (%s)\n🔁 Переместил: %d\n✅ Навёл порядок: %d\n⏱ Отчёты более 30 минут: %d\n\n",
			scaut.UserName, getDate(),
			b.getTimeReport(scaut.FirstTime), b.getTimeReport(scaut.TimeStart), scaut.TimeStart.Sub(scaut.FirstTime).String(),
			scaut.Moved, scaut.Images, scaut.Lateness,
		)
		b.bot.Send(MsgForAdmin)
	}

	// SCAUT Message
	sumHour := strings.Split(scaut.TimeStart.Sub(scaut.FirstTime).String(), "h")
	message := tgbotapi.NewMessage(msg.Chat.ID, "")

	message.Text = fmt.Sprintf("Смену завершил %s.c %s-%s (%s Часов)\n🔁Перемещения: %d \n✅Навёл порядок: %d \nИтого: %d", getDate(),
		b.getTimeReport(scaut.FirstTime), b.getTimeReport(scaut.TimeStart), sumHour[0],
		scaut.Moved, scaut.Images, scaut.Lateness,
	)
	b.AddStat(scaut, msg.From.ID)
	Scauts[msg.From.ID] = Scaut{}

	b.bot.Send(message)

}
func (b *Bot) CMDInfo(msg tgbotapi.Message) {
	b.muScauts.Lock()
	defer b.wg.Done()
	defer b.muScauts.Unlock()

	user := Scauts[msg.From.ID]
	message := tgbotapi.NewMessage(msg.Chat.ID, "")

	message.Text = fmt.Sprintf("Смену завершил %s.c %s-%s (%s)\n🔁Перемещения: %d \n✅Навёл порядок: %d \nИтого: %d",
		getDate(), b.getTimeReport(user.FirstTime), b.getTimeReport(user.TimeStart),
		user.TimeStart.Sub(user.FirstTime).String(),
		user.Moved, user.Images, user.Moved+user.Images,
	)
	b.bot.Send(message)
}

func (b *Bot) CMDStats(msg tgbotapi.Message) {
	b.muStats.Lock()
	defer b.wg.Done()
	defer b.muStats.Unlock()

	message := tgbotapi.NewMessage(msg.Chat.ID, "Общая Статистика!\n\n")
	for _, value := range StatScauts {
		if value.UserName != "" {
			message.Text += fmt.Sprintf("@%s\nПеремещений: %d\nОпозданий: %d\nВсего отработал: (%d)\n\n", value.UserName, value.SummerMuved, value.SummerLateness, value.SummerHour)
		}
	}
	b.bot.Send(message)

}

func (b *Bot) CMDRestats(msg tgbotapi.Message) {
	b.muStats.Lock()
	defer b.wg.Done()
	defer b.muStats.Unlock()
	for key, _ := range StatScauts {
		_, ok := StatScauts[key]
		if ok {
			StatScauts[key] = WendayScaut{}
		}
	}
	b.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Статистика удалена !"))
}
