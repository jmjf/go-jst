package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"go-slo/internal/jobStatus"
	modinit "go-slo/internal/jobStatus/infra/gormpg"
)

const (
	host     = "db"
	port     = 5432 // outside the container network it's 9432
	userName = "postgres"
	password = "postgres"
	dbName   = "go-slo"
)

type routeHandler struct {
	ctrl       *jobStatus.AddJobStatusCtrl
	baseLogger *slog.Logger
}

func (rh routeHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	logger := rh.baseLogger.With("route", request.URL.Path, "method", request.Method)
	if request.URL.Path == "/job-statuses" || request.URL.Path == "/job-statuses/" {
		switch request.Method {
		case http.MethodPost:
			rh.ctrl.Execute(response, request, logger)
		default:
			logger.Error("Not Implemented")
			response.WriteHeader(http.StatusNotImplemented)
		}
	} else {
		logger.Error("Unknown Route")
	}
}

func newLogger() *slog.Logger {
	hostName, err := os.Hostname()
	if err != nil {
		fmt.Printf("ERROR getting hostname: %v\n", err)
		hostName = "unknown"
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo, // TODO: get from env or command line
	})

	return slog.New(handler.WithAttrs([]slog.Attr{
		slog.String("applicationName", "go-slo"),
		slog.String("serviceName", "jobStatus"),
		slog.String("hostName", hostName),
		slog.Int64("pid", int64(os.Getpid())),
	}))
}

func main() {
	logger := newLogger()

	pgDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Etc/Utc", host, userName, password, dbName, port)

	dbRepo, _, ctrl, err := modinit.Init(pgDSN, logger)
	if err != nil {
		logger.Error("database connection failed", "err", err)
		panic(err)
	}
	defer dbRepo.Close()

	fmt.Println(" -- NewAddJobStatusController")
	rh := &routeHandler{
		ctrl:       ctrl,
		baseLogger: logger,
	}

	fmt.Println(" -- add routes")
	http.Handle("/job-statuses", rh)
	http.Handle("/job-statuses/", rh)

	fmt.Println(" -- start server")
	http.ListenAndServe(":9201", nil)

}
