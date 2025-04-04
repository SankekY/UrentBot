package main

import (
	"log"
	"urentBot/bot"
	"urentBot/config"

	"github.com/joho/godotenv"
)

func main() {
	Init()
	cofig := config.InitConfig()
	mybot := bot.NewBot(*cofig)
	log.Printf("Config Loaded!\n")
	mybot.Start()

}

func Init() {
	if err := godotenv.Load("../config/.env"); err != nil {
		log.Fatal(err)
	}
}
