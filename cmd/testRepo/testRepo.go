package main

//
// IMPORTANT NOTE
//
// This test requires the `JobStatusRepo`'s `add` method in `jobStatus/domain.go`
// to be renamed `Add` so it is exported. Normally, it should be hidden and available
// to the `jobStatus` module only, but this test needs access to it.
//

import (
	"database/sql"
	"fmt"
	"jobStatus"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	host     = "db"
	port     = 5432 // outside the container network it's 9432
	userName = "postgres"
	password = "postgres"
	dbName   = "go-slo"
)

var testData = []jobStatus.JobStatus{
	{
		ApplicationId:      "App1",
		JobId:              "Job1",
		JobStatusCode:      jobStatus.JobStatus_START,
		JobStatusTimestamp: time.Date(2023, 06, 02, 00, 52, 32, 0, time.UTC),
		BusinessDate:       time.Date(2023, 06, 01, 0, 0, 0, 0, time.UTC),
		RunId:              "",
		HostId:             "",
	},
	{
		ApplicationId:      "App1",
		JobId:              "Job1",
		JobStatusCode:      jobStatus.JobStatus_SUCCEED,
		JobStatusTimestamp: time.Date(2023, 06, 02, 01, 27, 59, 0, time.UTC),
		BusinessDate:       time.Date(2023, 06, 01, 0, 0, 0, 0, time.UTC),
		RunId:              "",
		HostId:             "",
	},
	{
		ApplicationId:      "App1",
		JobId:              "Job1",
		JobStatusCode:      jobStatus.JobStatus_SUCCEED,
		JobStatusTimestamp: time.Date(2023, 06, 03, 01, 27, 59, 0, time.UTC),
		BusinessDate:       time.Date(2023, 06, 02, 0, 0, 0, 0, time.UTC),
		RunId:              "",
		HostId:             "",
	},
	{
		ApplicationId:      "App2",
		JobId:              "Job980",
		JobStatusCode:      jobStatus.JobStatus_FAIL,
		JobStatusTimestamp: time.Date(2023, 06, 03, 00, 04, 12, 0, time.UTC),
		BusinessDate:       time.Date(2023, 06, 02, 0, 0, 0, 0, time.UTC),
		RunId:              "123456",
		HostId:             "localhost",
	},
}

func fmtData(data []jobStatus.JobStatus) (s string) {
	s = ""
	for _, js := range data {
		s += fmt.Sprintf("\t\t%s\n", js)
	}
	return s
}

func testRepo(testName string, repo jobStatus.JobStatusRepo, testData []jobStatus.JobStatus) {
	fmt.Printf("\n -- BEGIN TEST -- %s -- ", testName)

	fmt.Println("\n -- Add job status data")
	for _, js := range testData {
		fmt.Println("\tAdd", js)
		err := repo.Add(js)
		if err != nil {
			fmt.Println("\tAdd failed", err)
		}
	}

	fmt.Println("\n -- GetByJobId Job1")
	res, err := repo.GetByJobId("Job1")
	if err != nil {
		fmt.Println("\tError", err)
	}
	fmt.Printf("\tData\n%s", fmtData(res))

	fmt.Printf("\n -- GetByJobIdBusinessDate Job1 2023-06-02\n")
	res, err = repo.GetByJobIdBusinessDate("Job1", time.Date(2023, 06, 02, 0, 0, 0, 0, time.UTC))
	if err != nil {
		fmt.Println("\tError", err)
	}
	fmt.Printf("\tData\n%s", fmtData(res))

	fmt.Println("\n -- GetByJobId Job980")
	res, err = repo.GetByJobId("Job980")
	if err != nil {
		fmt.Println("\tError", err)
	}
	fmt.Printf("\tData\n%s", fmtData(res))

	fmt.Println("\n -- GetByJobId Job3 (doesn't exist)")
	res, err = repo.GetByJobId("Job3")
	if err != nil {
		fmt.Println("\tError", err)
	}
	fmt.Printf("\tData\n%s", fmtData(res))

	fmt.Printf(" -- END TEST -- %s\n", testName)
}

func main() {
	pgUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", userName, password, host, port, dbName)
	println(" -- Connect to ", pgUrl)
	db, err := sql.Open("pgx", pgUrl)
	if err != nil {
		fmt.Println("sql.Open failed", err)
		panic(err)
	}
	defer db.Close()

	fmt.Println("\n -- NewDbSqlRepo")
	dbSqlRepo := jobStatus.NewDbSqlPgRepo(db)

	testRepo("DbSqlPgRepo", dbSqlRepo, testData)

	fmt.Println("\n -- NewMemoryRepo")
	memoryRepo := jobStatus.NewMemoryRepo([]jobStatus.JobStatus{})

	testRepo("MemoryRepo", memoryRepo, testData)
}
