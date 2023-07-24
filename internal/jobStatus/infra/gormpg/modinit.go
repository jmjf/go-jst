package modinit

import (
	"log/slog"

	"go-slo/internal/jobStatus"
	"go-slo/internal/jobStatus/db/gormpg"
)

func Init(pgDSN string, logger *slog.Logger) (jobStatus.Repo, *jobStatus.UseCases, error) {
	logger.Info("create repo")
	dbRepo := gormpg.NewRepoDB(pgDSN)

	logger.Info("open database")
	err := dbRepo.Open()

	if err != nil {
		logger.Error("database connection failed", "err", err)
		return nil, nil, err
	}

	logger.Info("create usecases")
	uc := jobStatus.NewUseCases(dbRepo)

	return dbRepo, uc, nil
}
