package main

import (
	"log"
	"os"
	"time"
)

type Aula struct {
	dia        string
	inicio     time.Time
	fim        time.Time
	total      int
	disponivel int
	inscritos  int
}

func main() {
	err := configurarDB(os.Getenv("DB_PATH"))
	if err != nil {
		log.Fatal(err)
	}
	defer fecharDB()

	telegramChannel, err := configurarTelegram(os.Getenv("TELEGRAM_API_KEY"), os.Getenv("TELEGRAM_CHANNEL_ID"))
	if err != nil {
		log.Fatal(err)
	}

	atualizaAulaChannel := make(chan []Aula)
	go atualizaAulas(atualizaAulaChannel)

	for true {
		select {
		case telegramUpdate := <-telegramChannel:
			switch telegramUpdate.Message.Text {
			case "/vagas":
				err = enviaAulaFiltro(telegramUpdate.Message.Chat.ID, func(aula Aula) bool {
					return aula.disponivel > 0
				})
				if err != nil {
					log.Print(err)
				}
			case "/todos":
				err = enviaAulaFiltro(telegramUpdate.Message.Chat.ID, nil)
				if err != nil {
					log.Print(err)
				}
			}
		case aulas := <-atualizaAulaChannel:
			var diferencas map[int]Aula
			diferencas, err = obterDiferencasFromDB(aulas)

			for oldMessageID, aula := range diferencas {
				var newMessageID int
				newMessageID, err = updateAulaOnChannel(oldMessageID, aula)
				if err != nil {
					continue
				}

				err = atualizarAulaOnDB(oldMessageID, newMessageID, aula)
				if err != nil {
					log.Print(err)
				}
			}
		}
	}
}

func enviaAulaFiltro(chatID int64, filtro func(aula Aula) bool) (err error) {
	aulas, err := getAulasFromDB()
	if err != nil {
		return
	}

	for _, aula := range aulas {
		if filtro == nil || filtro(aula) {
			err = sendAulaToUser(chatID, aula)
			if err != nil {
				return
			}
		}
	}
	return
}

func atualizaAulas(atualizaAulasChannel chan<- []Aula) {
	for true {
		aulas, err := getAulasFromWeb()
		if err != nil {
			log.Print(err)
		} else {
			atualizaAulasChannel <- aulas
		}

		time.Sleep(10 * time.Minute)
	}
}
