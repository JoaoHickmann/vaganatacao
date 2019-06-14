package main

import (
	"database/sql"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/text/encoding/charmap"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
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
	db, err := configurarDB("/home/pedro/Projects/vaganatacao/horario.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	telegramChannel, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	esperaVagaChannel := make(chan []Aula)
	go esperaVaga(esperaVagaChannel)

	channelID, err := strconv.ParseInt(os.Getenv("TELEGRAM_CHANNEL_ID"), 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	for true {
		select {
		case telegramUpdate := <-telegramChannel:
			switch telegramUpdate.ChannelPost.Text {
			case "/vagas":

			case "/todos":
				html, err := getHtml()
				if err != nil {
					log.Fatal(err)
				}

				msg := ""
				aulas, err := getAulas(html)
				for _, aula := range aulas {
					msg = msg + fmt.Sprintf("Dia: %s\nInicio: %s\nFim: %s\nTotal: %d\nDisponivel: %d\nInscritos: %d\n\n", aula.dia,
						aula.inicio.Format("15:04"), aula.fim.Format("15:04"), aula.total, aula.disponivel, aula.inscritos)
				}

			}
		case aulasDisponiveis := <-esperaVagaChannel:
			msg := ""
			for _, aula := range aulasDisponiveis {
				msg = msg + fmt.Sprintf("Dia: %s\nInicio: %s\nFim: %s\nTotal: %d\nDisponivel: %d\nInscritos: %d\n\n", aula.dia,
					aula.inicio.Format("15:04"), aula.fim.Format("15:04"), aula.total, aula.disponivel, aula.inscritos)
			}

			_, err = bot.Send(tgbotapi.NewMessage(channelID, msg))
			if err != nil {
				log.Print(err)
			}
		}
	}
}

func configurarDB(dataSourceName string) (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return
	}

	sqlCreate := `
		CREATE TABLE IF NOT EXISTS aulas
		(
			seq        INTEGER PRIMARY KEY AUTOINCREMENT,
			dia        TEXT    NOT NULL,
			inicio     DATE    NOT NULL,
			fim        DATE    NOT NULL,
			total      INTEGER NOT NULL,
			disponivel INTEGER NOT NULL,
			inscritos  INTEGER NOT NULL
		)
	`
	_, errCreate := db.Exec(sqlCreate)
	if errCreate != nil {
		log.Print(errCreate)
	}

	return
}

func trataAulas(canal chan<- []Aula, aulas []Aula) {
	aulasDisponiveis := []Aula{}
	for _, aula := range aulas {
		if aula.disponivel > 0 {
			aulasDisponiveis = append(aulasDisponiveis, aula)
		}
	}
	canal <- aulasDisponiveis
}

func esperaVaga(canal chan<- []Aula) {
	for true {
		html, err := getHtml()
		if err != nil {
			log.Print(err)
			time.Sleep(10 * time.Minute)
			continue
		}

		aulas, err := getAulas(html)
		if err != nil {
			log.Print(err)
			time.Sleep(10 * time.Minute)
			continue
		}

		trataAulas(canal, aulas)

		time.Sleep(10 * time.Minute)
	}
}

func getHtml() (html string, err error) {
	response, err := http.Get("https://www.univates.br/esporte-e-saude/vagas")
	if err != nil {
		return
	}
	defer response.Body.Close()

	utf8Html, err := ioutil.ReadAll(charmap.ISO8859_1.NewDecoder().Reader(response.Body))
	if err != nil {
		return
	}

	html = string(utf8Html)
	return
}

func getAulas(html string) (aulas []Aula, err error) {
	regexNatatacao, err := regexp.Compile(`<div class="item">\s*?<div class="item-plus icon-plus-circled">Aprendizagem ADULTO</div>\s*?<div class="item-more">([\s\S]+?)</div>\s*?</div>`)
	if err != nil {
		return
	}
	regexAulas, err := regexp.Compile(`<strong>Horário: </strong>(.+?)<br />\s*?<strong>Total Vagas: </strong>(.+?)<br />\s*?<strong>Vagas Disponíveis: </strong>(.+?)<br />\s*?<strong>Inscritos: </strong>(.+?)<br />\s*?<hr />`)
	if err != nil {
		return
	}

	htmlAulas := regexNatatacao.FindStringSubmatch(html)[1]
	aulasString := regexAulas.FindAllStringSubmatch(htmlAulas, -1)
	for _, aulaString := range aulasString {
		var aula Aula
		aula, err = getAula(aulaString)
		if err != nil {
			return
		}
		aulas = append(aulas, aula)
	}
	return
}

func getAula(aulaString []string) (aula Aula, err error) {
	regexHorario, err := regexp.Compile(`(\d\d:\d\d)-(\d\d:\d\d) - (.+)`)
	if err != nil {
		return
	}
	horarioDetalhado := regexHorario.FindStringSubmatch(aulaString[1])

	dia := horarioDetalhado[3]

	inicio, err := time.Parse("15:04", horarioDetalhado[1])
	if err != nil {
		return
	}

	fim, err := time.Parse("15:04", horarioDetalhado[2])
	if err != nil {
		return
	}

	regexTotal := regexp.MustCompile(`(\d+) / Hor`)
	totalString := regexTotal.FindStringSubmatch(aulaString[2])[1]
	total, err := strconv.Atoi(totalString)
	if err != nil {
		return
	}

	disponivel, err := strconv.Atoi(aulaString[3])
	if err != nil {
		return
	}

	inscritos, err := strconv.Atoi(aulaString[4])
	if err != nil {
		return
	}

	aula = Aula{
		dia:        dia,
		inicio:     inicio,
		fim:        fim,
		total:      total,
		disponivel: disponivel,
		inscritos:  inscritos,
	}
	return
}
