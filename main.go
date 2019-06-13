package main

import (
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"io/ioutil"
	"log"
	"net/http"
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
	html, err := getHtml()
	if err != nil {
		log.Fatal(err)
	}

	aulas, err := getAulas(html)
	for _, aula := range aulas {
		if aula.inicio.Hour() == 18 && aula.inicio.Minute() == 0 {
			fmt.Printf("Dia: %s\nInicio: %s\nFim: %s\nTotal: %d\nDisponivel: %d\nInscritos: %d\n\n", aula.dia,
				aula.inicio.Format("15:04"), aula.fim.Format("15:04"), aula.total, aula.disponivel, aula.inscritos)
		}
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
