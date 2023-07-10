package modinit

import (
	"fmt"
	"log/slog"

	"go-slo/internal/jobStatus"
	dbpg "go-slo/internal/jobStatus/db_sqlpgx"
)

func Init(pgUrl string, logger *slog.Logger) (jobStatus.Repo, *jobStatus.AddJobStatusUC, *jobStatus.AddJobStatusCtrl, error) {

	fmt.Println(" -- NewRepoDb")
	dbRepo := dbpg.NewRepoDB(pgUrl)

	fmt.Println(" -- Open database connection")
	err := dbRepo.Open()
	if err != nil {
		logger.Error("database connection failed", "err", err)
		return nil, nil, nil, err
	}

	fmt.Println(" -- NewAddJobStatusUC")
	uc := jobStatus.NewAddJobStatusUC(dbRepo)

	fmt.Println(" -- NewAddJobStatusController")
	ctrl := jobStatus.NewAddJobStatusCtrl(uc)

	return dbRepo, uc, ctrl, nil
}
