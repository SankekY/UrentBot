package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	myBot := NewTeleBot("1556866128:AAGt8VJ-nfAGMLYf6K8UZuHDDhJeHoyoltQ")
	myBot.Start()
}

type TelegramBot struct {
	tgBot *tgbotapi.BotAPI
}

func NewTeleBot(token string) *TelegramBot {
	bot, _ := tgbotapi.NewBotAPI(token)
	return &TelegramBot{
		tgBot: bot,
	}
}

/*
TODO: Проверка на РГЛ или другие звания если в чат для отчётов скинули не отчёт удалять если сообщение от РГЛ и тд. Оставить !
*/

type User struct {
	UserName string
	Moved    int
	Images   int
}

var Users = make(map[int64]User)

func (bot *TelegramBot) Start() {
	u := tgbotapi.NewUpdate(0)
	updates := bot.tgBot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message.Text == "" && update.Message.Photo == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		if update.Message.IsCommand() {
			if update.Message.Chat.Type != "private" {
				continue
			}
			switch update.Message.Command() {
			case "start":
				msg.Text = "/info - Ифнормация по перемещениям и фото !\n /reset - Сбросить счётчик"
				bot.tgBot.Send(msg)
				continue
			case "info":
				user := Users[update.Message.From.ID]
				sum := user.Images + user.Moved
				_, month, day := time.Now().Date()
				date := fmt.Sprintf("%d.%d", day, month)
				if day < 10 && month < 10 {
					date = fmt.Sprintf("0%d.0%d", day, month)
				} else if day < 10 && month > 10 {
					date = fmt.Sprintf("0%d.%d", day, month)
				} else if day > 10 && month < 10 {
					date = fmt.Sprintf("%d.0%d", day, month)
				}

				msg.Text = fmt.Sprintf("Смену завершил %s.c 00:00-00:00 (0 Часов)\n🔁Перемещения: %d \n✅Навёл порядок: %d \nИтого: %d", date, user.Moved, user.Images, sum)
				bot.tgBot.Send(msg)
				continue
			case "reset":
				Users[update.Message.From.ID] = User{}
				user := Users[update.Message.From.ID]
				sum := user.Images + user.Moved
				msg.Text = fmt.Sprintf("Смену завершил 00.00.c 00:00-00:00 (0 Часов)\n🔁Перемещения: %d \n✅Навёл порядок: %d \nИтого: %d", user.Moved, user.Images, sum)
				bot.tgBot.Send(msg)
				continue
			case "all":
				msg.Text = fmt.Sprintln(Users)
				bot.tgBot.Send(msg)
			}

		}
		user := Users[update.Message.From.ID]
		user.UserName = update.Message.From.UserName
		if update.Message.Photo != nil {
			user.Images += 1
			Users[update.Message.From.ID] = user
			continue
		}
		mov, err := SerhMoved(update.Message.Text)
		if err != nil {
			continue
		}
		user.Moved += mov
		Users[update.Message.From.ID] = user

	}

}

func SerhMoved(msg string) (int, error) {
	arrText := strings.Split(msg, "\n")
	NewArr := strings.Split(arrText[len(arrText)-1], " ")
	if len(NewArr) < 2 {
		return 0, fmt.Errorf("")
	}
	return strconv.Atoi(NewArr[1])
}
