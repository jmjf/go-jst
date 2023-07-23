package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	jshttp "go-slo/internal/jobStatus/http"
	modinit "go-slo/internal/jobStatus/infra/gormpg"
	"go-slo/internal/middleware"
)

const (
	host     = "db"
	port     = 5432 // outside the container network it's 9432
	userName = "postgres"
	password = "postgres"
	dbName   = "go-slo"
)

func logHandler(rootLogger *slog.Logger, tag string) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		logger := rootLogger.With("tag", tag, "route", req.URL.Path, "method", req.Method)
		logger.Info("received", "urlRawPath", req.URL.RawPath, "urlString", req.URL.String())
	})
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

	logger.Info("initialize application")
	pgDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Etc/Utc", host, userName, password, dbName, port)
	dbRepo, _, ctrl, err := modinit.Init(pgDSN, logger)
	if err != nil {
		logger.Error("database connection failed", "err", err)
		panic(err)
	}
	defer dbRepo.Close()

	logger.Info("build mux")
	apiMux := http.NewServeMux()
	mux := http.NewServeMux()
	logRequestMw := middleware.BuildReqLoggerMw(logger)

	apiMux.Handle("/job-statuses", jshttp.Handler(logger, ctrl))
	apiMux.Handle("/job-statuses/", jshttp.Handler(logger, ctrl))
	mux.Handle("/api/", http.StripPrefix("/api", middleware.AddRequestId(logRequestMw(apiMux))))
	mux.Handle("/", logHandler(logger, "/"))

	logger.Info("start server")
	http.ListenAndServe(":9201", mux)

}
