package main

import (
	"database/sql"
	"fmt"
	"jobStatus"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	host     = "db"
	port     = 5432 // outside the container network it's 9432
	userName = "postgres"
	password = "postgres"
	dbName   = "go-slo"
)

type routeHandler struct {
	ctrl jobStatus.JobStatusCtrl
}

func (rh *routeHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/job-statuses" || request.URL.Path == "/job-statuses/" {
		switch request.Method {
		case http.MethodPost:
			rh.ctrl.AddJobStatus(response, request)
		default:
			response.WriteHeader(http.StatusNotImplemented)
		}
	}
}

func main() {
	pgUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", userName, password, host, port, dbName)
	fmt.Printf(" -- Connect to %s\n", pgUrl)
	db, err := sql.Open("pgx", pgUrl)
	if err != nil {
		fmt.Printf("sql.Open failed %s\n", err)
		panic(err)
	}
	defer db.Close()

	fmt.Println(" -- NewDbSqlRepo")
	dbSqlRepo := jobStatus.NewDbSqlPgRepo(db)

	fmt.Println(" -- NewJobStatusUC")
	uc := jobStatus.NewJobStatusUC(dbSqlRepo)

	fmt.Println(" -- NewJobStatusController")
	rh := &routeHandler{
		ctrl: jobStatus.NewJobStatusController(uc),
	}

	fmt.Println(" -- add routes")
	http.Handle("/job-statuses", rh)
	http.Handle("/job-statuses/", rh)

	fmt.Println(" -- start server")
	http.ListenAndServe(":9201", nil)

}
