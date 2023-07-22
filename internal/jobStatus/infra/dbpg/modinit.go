package modinit

import (
	"fmt"
	"log/slog"

	"go-slo/internal/jobStatus"
	dbpg "go-slo/internal/jobStatus/db_sqlpgx"
)

func Init(pgUrl string, logger *slog.Logger) (jobStatus.Repo, *jobStatus.AddJobStatusUC, *jobStatus.GetByQueryUC, *jobStatus.AddJobStatusCtrl, *jobStatus.GetByQueryCtrl, error) {

	fmt.Println(" -- NewRepoDb")
	dbRepo := dbpg.NewRepoDB(pgUrl)

	fmt.Println(" -- Open database connection")
	err := dbRepo.Open()
	if err != nil {
		logger.Error("database connection failed", "err", err)
		return nil, nil, nil, nil, nil, err
	}

	fmt.Println(" -- NewAddJobStatusUC")
	addUC := jobStatus.NewAddJobStatusUC(dbRepo)

	fmt.Println(" -- NewAddJobStatusController")
	addCtrl := jobStatus.NewAddJobStatusCtrl(addUC)

	fmt.Println(" -- NewGetByQueryUC")
	getUC := jobStatus.NewGetByQueryUC(dbRepo)

	fmt.Println(" -- NewGetByQueryController")
	queryCtrl := jobStatus.NewGetByQueryCtrl(getUC)

	return dbRepo, addUC, getUC, addCtrl, queryCtrl, nil
}
