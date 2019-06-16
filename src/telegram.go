package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
)

var bot *tgbotapi.BotAPI
var channelID int64

func configurarTelegram(token string, _channelID string) (updatesChannel tgbotapi.UpdatesChannel, err error) {
	channelID, err = strconv.ParseInt(_channelID, 10, 64)
	if err != nil {
		return
	}

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

	msg := fmt.Sprintf("Dia: %s\nInicio: %s\nFim: %s\nTotal: %d\nDisponivel: %d\nInscritos: %d\n\n",
		aula.dia, aula.inicio.Format("15:04"), aula.fim.Format("15:04"), aula.total, aula.disponivel, aula.inscritos)

	message, err := bot.Send(tgbotapi.NewMessage(channelID, msg))
	if err != nil {
		return
	}

	newMessageID = message.MessageID

	return
}

func sendAulaToUser(chatID int64, aula Aula) (err error) {
	msg := fmt.Sprintf("Dia: %s\nInicio: %s\nFim: %s\nTotal: %d\nDisponivel: %d\nInscritos: %d\n\n",
		aula.dia, aula.inicio.Format("15:04"), aula.fim.Format("15:04"), aula.total, aula.disponivel, aula.inscritos)
	_, err = bot.Send(tgbotapi.NewMessage(chatID, msg))
	return
}
