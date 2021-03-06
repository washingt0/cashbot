package main

import (
	"context"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/washingt0/cashbot/database"
	"github.com/washingt0/cashbot/handler"
)

func main() {
	token := ""
	if val, set := os.LookupEnv("CASHBOT_API_TOKEN"); set {
		token = val
	} else {
		log.Fatal("No API token was supplied")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	debug := false
	if _, set := os.LookupEnv("CASHBOT_DEBUG"); set {
		debug = true
	}
	bot.Debug = debug

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		ctx, cancel := context.WithCancel(context.TODO())

		client, err := database.ConnectMongo(ctx)
		if err != nil {
			log.Fatal(err)
		}

		msg, isFile, err := handler.IncomingMessageHandler(update.Message, client)
		if err != nil {
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
		}

		msg.ReplyToMessageID = update.Message.MessageID

		if isFile {
			doc := tgbotapi.NewDocumentUpload(update.Message.Chat.ID, msg.Text)
			bot.Send(doc)
		} else {
			bot.Send(msg)
		}

		client.Disconnect(ctx)
		cancel()
	}
}
