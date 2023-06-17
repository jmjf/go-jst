package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	host     = "db"
	port     = 5432 // outside the container network, 9432
	userName = "postgres"
	password = "postgres"
	dbName   = "gojst"
)

func main() {
	pgUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", userName, password, host, port, dbName)
	println("Connect to ", pgUrl)
	db, err := sql.Open("pgx", pgUrl)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	fmt.Println("Query")

	var jobId int
	var statusCode string
	var statusTimestamp, businessDate time.Time

	rows, err := db.Query(`SELECT * FROM "JobStatus"`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		panic(err)
	}
	fmt.Println(cols)
	for rows.Next() {

		err = rows.Scan(&jobId, &statusCode, &statusTimestamp, &businessDate)
		if err != nil {
			panic(err)
		}
		fmt.Println(jobId, "|", statusCode, "|",
			statusTimestamp.Format("2006-01-02 15:04:05.00 -0700"), "|",
			businessDate.Format("2006-01-02"))
	}

	fmt.Println("Successful connection")
}
