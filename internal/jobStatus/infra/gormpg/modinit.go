package modinit

import (
	"fmt"
	"log/slog"

	"go-slo/internal/jobStatus"
	gormpg "go-slo/internal/jobStatus/db_gormpg"
)

func Init(pgDSN string, logger *slog.Logger) (jobStatus.Repo, *jobStatus.AddJobStatusUC, *jobStatus.AddJobStatusCtrl, error) {
	fmt.Println(" -- NewRepoDb")
	dbRepo := gormpg.NewRepoDB(pgDSN)

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
