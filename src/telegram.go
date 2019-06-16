package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var bot *tgbotapi.BotAPI
var channelID int64

func configurarTelegram(token string, _channelID int64) (updatesChannel tgbotapi.UpdatesChannel, err error) {
	channelID = _channelID

	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		return
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updatesChannel, err = bot.GetUpdatesChan(u)
	return
}

func updateAulaOnChannel(oldMessageID int, aula Aula) (newMessageID int, err error) {
	if oldMessageID >= 0 {
		_, err = bot.Send(tgbotapi.NewDeleteMessage(channelID, oldMessageID))
		if err != nil {
			return
		}
	}

	message, err := bot.Send(tgbotapi.NewMessage(channelID, aula.toString()))
	if err != nil {
		return
	}

	newMessageID = message.MessageID
	return
}

func sendAulaToUser(chatID int64, aula Aula) (err error) {
	_, err = bot.Send(tgbotapi.NewMessage(chatID, aula.toString()))
	return
}
