package ADDS

import (
	"database/sql"
	"encoding/csv"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"os"
)

var CSVRequiredFields = []string{"ident", "type", "name", "latitude_deg", "longitude_deg", "continent", "iso_country"}

// Imports database from http://ourairports.com/data/ file.
func ImportCSVToNewSQLite(csvFile, dbFile string) error {

	// Open SQLite database.
	os.Remove(dbFile)
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create new table.
	sqlStmt := `
	create table airports (id integer not null primary key, ident text, type text, name text, latitude_deg real, longitude_deg real, continent text, iso_country text);
	delete from airports;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return err
	}

	// Prepare insert statement.
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert into airports(ident, type, name, latitude_deg, longitude_deg, continent, iso_country) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Open CSV file and start parsing.
	f, err := os.Open(csvFile)
	if err != nil {
		return err
	}
	defer f.Close()

	colLabels := make(map[string]int, 0)
	reqLen := 0

	// Start reading in the CSV.
	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if len(colLabels) == 0 {
			// Load in the labels.
			for i := 0; i < len(record); i++ {
				colLabels[record[i]] = i
			}
			// Make sure we have the required fields.
			for _, v := range CSVRequiredFields {
				if fieldPos, ok := colLabels[v]; !ok {
					return errors.New("missing required fields.")
				} else {
					if fieldPos > reqLen {
						reqLen = fieldPos // Track the maximum field position in the required fields.
					}
				}
			}
			reqLen++
			continue
		}

		// Extract ident, type, name, latitude_deg, longitude_deg, continent, iso_country.
		if len(record) < reqLen {
			continue
		}

		vals := make([]interface{}, 0)
		for _, field := range CSVRequiredFields {
			vals = append(vals, record[colLabels[field]])
		}

		_, err = stmt.Exec(vals...)
		if err != nil {
			return err
		}

	}

	tx.Commit()

	return nil
}
