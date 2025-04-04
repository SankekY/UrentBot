package bot

import (
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
	🔻 ___ *Scaut Mode* ___ 🔻
	│
	│ 📌 */info* → Текущая статистика 
	│ (без сброса смены)             
	│                                   
	│ 📝 */report* → Завершить смену и  
	│ сгенерировать отчёт            
	│                                   
	│ 🔔 */sub* → Авто-напоминание 
	│ (пришлёт отчёт в ЛС, если      
	│ не завершить смену за 1.5 часа)
	│                                  
	🔺 ____________________ 🔺
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
			return
		} else {
			b.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Не понятная комманда напишите \n/start -Для просмтора комманд"))
		}

	}
}

func (b *Bot) CMDStart(msg tgbotapi.Message) {
	defer b.wg.Done()

	if !b.cfg.Admins[msg.From.ID] {
		b.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, AwaitMessage["start_scaut"]))
		return
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

	message := tgbotapi.NewMessage(msg.Chat.ID, "")

	message.Text += b.RGLStats(Scauts)

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
		MsgForAdmin.Text = b.GenerateReportRGL(scaut)
		b.bot.Send(MsgForAdmin)

	}

	// SCAUT Message

	message := tgbotapi.NewMessage(msg.Chat.ID, "")
	message.Text = b.GenerateReportScaut(scaut)
	b.bot.Send(message)

	b.AddStat(scaut, msg.From.ID)
	Scauts[msg.From.ID] = Scaut{}
}

func (b *Bot) CMDInfo(msg tgbotapi.Message) {
	b.muScauts.Lock()
	defer b.wg.Done()
	defer b.muScauts.Unlock()

	user := Scauts[msg.From.ID]
	message := tgbotapi.NewMessage(msg.Chat.ID, "")
	message.Text = b.GenerateReportScaut(user)
	b.bot.Send(message)
}

func (b *Bot) CMDStats(msg tgbotapi.Message) {
	b.muStats.Lock()
	defer b.wg.Done()
	defer b.muStats.Unlock()

	message := tgbotapi.NewMessage(msg.Chat.ID, "")

	message.Text += GenerateStats(StatScauts)

	b.bot.Send(message)

}

func (b *Bot) CMDRestats(msg tgbotapi.Message) {
	b.muStats.Lock()
	defer b.wg.Done()
	defer b.muStats.Unlock()

	StatScauts = make(map[int64]WendayScaut)

	b.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Статистика удалена !"))
}
