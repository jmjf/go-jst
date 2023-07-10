package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"go-slo/internal/jobStatus"
	modinit "go-slo/internal/jobStatus/infra/dbpg"

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
	handlerOpts := slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo, // TODO: get from env or command line
	}
	handler := slog.NewJSONHandler(os.Stdout, &handlerOpts)

	hostName, err := os.Hostname()
	if err != nil {
		fmt.Printf("ERROR getting hostname: %v\n", err)
	}

	return slog.New(handler.WithAttrs([]slog.Attr{
		slog.String("applicationName", "go-slo"),
		slog.String("serviceName", "jobStatus"),
		slog.String("hostName", hostName),
		slog.Int64("pid", int64(os.Getpid())),
	}))
}

func main() {
	logger := newLogger()

	pgUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", userName, password, host, port, dbName)
	dbRepo, _, ctrl, err := modinit.Init(pgUrl, logger)
	if err != nil {
		logger.Error("init failed", "err", err)
		panic(err)
	}
	defer dbRepo.Close()

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
