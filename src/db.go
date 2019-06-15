package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

var db *sql.DB

func configurarDB(dataSourceName string) (err error) {
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
	_, err = db.Exec(sqlCreate)
	if err != nil {
		return
	}

	return
}

func fecharDB() {
	err := db.Close()
	if err != nil {
		log.Print(err)
	}
}

func inserirAula(aula Aula) (err error) {
	_, err = db.Exec(`
		INSERT INTO aulas (dia, inicio, fim, total, disponivel, inscritos)
		VALUES (?, ?, ?, ?, ?, ?)
	`, aula.dia, aula.inicio, aula.fim, aula.total, aula.disponivel, aula.inscritos)
	return
}

func getUltimasDuasAulas() (aulas [][]Aula, err error) {
	rowsHorarios, err := db.Query(`
		SELECT dia, inicio, fim
		FROM aulas
		GROUP BY dia, inicio, fim
	`)
	if err != nil {
		return
	}
	defer rowsHorarios.Close()

	var dia string
	var inicio time.Time
	var fim time.Time
	for rowsHorarios.Next() {
		err = rowsHorarios.Scan(&dia, &inicio, &fim)
		if err != nil {
			return
		}

		aulas = append(aulas, []Aula{})

		var rowsVagas *sql.Rows
		rowsVagas, err = db.Query(`
			SELECT total, disponivel, inscritos
			FROM aulas
			WHERE dia = ?
			  AND inicio = ?
			  AND fim = ?
			ORDER BY seq DESC
			LIMIT 2
		`, dia, inicio, fim)
		if err != nil {
			return
		}
		defer rowsVagas.Close()

		var total int
		var disponivel int
		var inscritos int
		for rowsVagas.Next() {
			err = rowsVagas.Scan(&total, &disponivel, &inscritos)
			if err != nil {
				return
			}

			ultimaAula := len(aulas) - 1
			aulas[ultimaAula] = append(aulas[ultimaAula], Aula{
				dia,
				inicio,
				fim,
				total,
				disponivel,
				inscritos,
			})
		}
	}
	return
}
