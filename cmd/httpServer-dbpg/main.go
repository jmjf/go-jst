package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	jshttp "go-slo/internal/jobStatus/http"
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

func logHandler(rootLogger *slog.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		logger := rootLogger.With("route", req.URL.Path, "method", req.Method)
		logger.Info("received", "urlRawPath", req.URL.RawPath, "urlString", req.URL.String())
	})
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

	fmt.Println(" -- initialize app")
	pgUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", userName, password, host, port, dbName)
	dbRepo, _, addCtrl, err := modinit.Init(pgUrl, logger)
	if err != nil {
		logger.Error("init failed", "err", err)
		panic(err)
	}
	defer dbRepo.Close()

	fmt.Println(" -- build mux")
	// apiMux := http.NewServeMux()
	subMux := http.NewServeMux()

	subMux.Handle("/api/job-statuses", jshttp.Handler(logger, addCtrl))
	subMux.Handle("/api/job-statuses/", jshttp.Handler(logger, addCtrl))
	subMux.Handle("/", logHandler(logger))

	fmt.Println(" -- start server")
	http.ListenAndServe(":9201", subMux)
}
