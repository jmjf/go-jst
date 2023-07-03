package main

import (
	"fmt"
	"jobStatus"
	"log/slog"
	"net/http"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

const (
	host     = "db"
	port     = 5432 // outside the container network it's 9432
	userName = "postgres"
	password = "postgres"
	dbName   = "go-slo"
)

type routeHandler struct {
	ctrl       jobStatus.JobStatusCtrl
	baseLogger *slog.Logger
}

func (rh routeHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	logger := rh.baseLogger.With("route", request.URL.Path, "method", request.Method)
	if request.URL.Path == "/job-statuses" || request.URL.Path == "/job-statuses/" {
		switch request.Method {
		case http.MethodPost:
			rh.ctrl.AddJobStatus(response, request, logger)
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

	pgDsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Etc/Utc", host, userName, password, dbName, port)
	fmt.Printf(" -- Connect to %s\n", pgDsn)
	db, err := gorm.Open(postgres.Open(pgDsn), &gorm.Config{
		TranslateError: true,
		Logger:         gormLogger.Default.LogMode(gormLogger.Silent),
		// Logger: logger, // doesn't work because gorm's logger interface is different; will need to translate
		NowFunc: func() time.Time { return time.Now().UTC() }, // ensure times are UTC
		// PrepareStmt: true // cache prepared statements for SQL; need to investigate how this works before turning on
	})
	if err != nil {
		logger.Error("gorm.Open failed", "err", err)
		panic(err)
	}
	// gorm doesn't have a Close()

	fmt.Println(" -- NewDbSqlRepo")
	gormRepo := jobStatus.NewGormPgRepo(db)

	fmt.Println(" -- NewJobStatusUC")
	uc := jobStatus.NewJobStatusUC(gormRepo)

	fmt.Println(" -- NewJobStatusController")
	rh := &routeHandler{
		ctrl:       jobStatus.NewJobStatusCtrl(uc),
		baseLogger: logger,
	}

	fmt.Println(" -- add routes")
	http.Handle("/job-statuses", rh)
	http.Handle("/job-statuses/", rh)

	fmt.Println(" -- start server")
	http.ListenAndServe(":9201", nil)

}
