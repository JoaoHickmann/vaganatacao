package main

import (
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/text/encoding/charmap"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
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
	err := configurarDB(os.Getenv("DB_PATH"))
	if err != nil {
		log.Fatal(err)
	}
	defer fecharDB()

	err = configurarTelegram(os.Getenv("TELEGRAM_API_KEY"), os.Getenv("TELEGRAM_CHANNEL_ID"))
	if err != nil {
		log.Fatal(err)
	}

	telegramChannel, err := getTelegramUpdateChannel()
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
				var ultimasDuasAulas [][]Aula
				ultimasDuasAulas, err = getUltimasDuasAulas()
				if err != nil {
					log.Print(err)
					continue
				}

				for _, duasAulas := range ultimasDuasAulas {
					if len(duasAulas) > 0 && duasAulas[0].disponivel > 0 {
						err = sendAulaToUser(telegramUpdate.Message.Chat.ID, duasAulas[0])
						if err != nil {
							log.Print(err)
						}
					}
				}
			case "/todos":
				var ultimasDuasAulas [][]Aula
				ultimasDuasAulas, err = getUltimasDuasAulas()
				if err != nil {
					log.Print(err)
					continue
				}

				for _, duasAulas := range ultimasDuasAulas {
					if len(duasAulas) > 0 {
						err = sendAulaToUser(telegramUpdate.Message.Chat.ID, duasAulas[0])
						if err != nil {
							log.Print(err)
						}
					}
				}
			}
		case aulas := <-atualizaAulaChannel:
			for _, aula := range aulas {
				err = inserirAula(aula)
				if err != nil {
					log.Print(err)
				}
			}

			var ultimasDuasAulas [][]Aula
			ultimasDuasAulas, err = getUltimasDuasAulas()
			if err != nil {
				log.Print(err)
				continue
			}

			for _, duasAulas := range ultimasDuasAulas {
				if len(duasAulas) == 1 || (len(duasAulas) == 2 && !reflect.DeepEqual(duasAulas[0], duasAulas[1])) {
					err = sendAulaToChannel(duasAulas[0])
					if err != nil {
						log.Print(err)
						continue
					}
				}
			}
		}
	}
}

func atualizaAulas(atualizaAulasChannel chan<- []Aula) {
	for true {
		aulas, err := getAulasAtualizadas()
		if err != nil {
			log.Print(err)
			time.Sleep(10 * time.Minute)
			continue
		}

		atualizaAulasChannel <- aulas

		time.Sleep(10 * time.Minute)
	}
}

func getAulasAtualizadas() (aulas []Aula, err error) {
	var html string
	html, err = getHtml()
	if err != nil {
		return
	}

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
