package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
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
			messageID  INTEGER PRIMARY KEY AUTOINCREMENT,
			data       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			dia        TEXT      NOT NULL,
			inicio     TIMESTAMP NOT NULL,
			fim        TIMESTAMP NOT NULL,
			total      INTEGER   NOT NULL,
			disponivel INTEGER   NOT NULL,
			inscritos  INTEGER   NOT NULL
		)
	`
	_, err = db.Exec(sqlCreate)
	return
}

func fecharDB() {
	err := db.Close()
	if err != nil {
		log.Print(err)
	}
}

func obterDiferencasFromDB(aulas []Aula) (diferencas map[int]Aula, err error) {
	diferencas = make(map[int]Aula)

	var messageID, total, disponivel, incritos int
	for index, aula := range aulas {
		row := db.QueryRow(`
			SELECT messageID, total, disponivel, inscritos
			FROM aulas
			WHERE dia = ?
			  AND inicio = ?
			  AND fim = ?
		`, aula.dia, aula.inicio, aula.fim)
		err = row.Scan(&messageID, &total, &disponivel, &incritos)
		if err != nil && err != sql.ErrNoRows {
			return
		} else if err == sql.ErrNoRows {
			messageID = index - 10000
		}

		if err == sql.ErrNoRows || aula.total != total || aula.disponivel != disponivel || aula.inscritos != incritos {
			diferencas[messageID] = aula
		}
	}
	return
}

func atualizarAulaOnDB(oldMessageID int, newMessageID int, aula Aula) (err error) {
	_, err = db.Exec(`
		DELETE FROM aulas
		WHERE messageID = ?
	`, oldMessageID)
	if err != nil {
		return
	}

	_, err = db.Exec(`
		INSERT INTO aulas (messageID, dia, inicio, fim, total, disponivel, inscritos)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, newMessageID, aula.dia, aula.inicio, aula.fim, aula.total, aula.disponivel, aula.inscritos)
	return
}

func getAulasFromDB() (aulas []Aula, err error) {
	rows, err := db.Query(`
		SELECT dia, inicio, fim, total, disponivel, inscritos
		FROM aulas
		ORDER BY dia, inicio
	`)
	if err != nil {
		return
	}

	for rows.Next() {
		var aula Aula
		err = rows.Scan(&aula.dia, &aula.inicio, &aula.fim, &aula.total, &aula.disponivel, &aula.inscritos)
		if err != nil {
			return
		}
		aulas = append(aulas, aula)
	}
	return
}
