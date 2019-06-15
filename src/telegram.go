package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var bot *tgbotapi.BotAPI
var channelID string

func configurarTelegram(token string, _channelID string) (err error) {
	channelID = _channelID
	bot, err = tgbotapi.NewBotAPI(token)
	return
}

func getTelegramUpdateChannel() (updatesChannel tgbotapi.UpdatesChannel, err error) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updatesChannel, err = bot.GetUpdatesChan(u)
	return
}

func sendAulaToChannel(aula Aula) (err error) {
	msg := fmt.Sprintf("Dia: %s\nInicio: %s\nFim: %s\nTotal: %d\nDisponivel: %d\nInscritos: %d\n\n",
		aula.dia, aula.inicio.Format("15:04"), aula.fim.Format("15:04"), aula.total, aula.disponivel, aula.inscritos)
	_, err = bot.Send(tgbotapi.NewMessageToChannel(channelID, msg))
	return
}

func sendAulaToUser(chatID int64, aula Aula) (err error) {
	msg := fmt.Sprintf("Dia: %s\nInicio: %s\nFim: %s\nTotal: %d\nDisponivel: %d\nInscritos: %d\n\n",
		aula.dia, aula.inicio.Format("15:04"), aula.fim.Format("15:04"), aula.total, aula.disponivel, aula.inscritos)
	_, err = bot.Send(tgbotapi.NewMessage(chatID, msg))
	return
}
