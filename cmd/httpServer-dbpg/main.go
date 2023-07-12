package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	jshttp "go-slo/internal/jobStatus/http"
	modinit "go-slo/internal/jobStatus/infra/dbpg"
	"go-slo/lib/middleware"

	_ "github.com/jackc/pgx/v5/stdlib"
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

func newLogger(appName string, svcName string) *slog.Logger {
	handlerOpts := slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo, // TODO: get from env or command line
	}
	handler := slog.NewJSONHandler(os.Stdout, &handlerOpts)

	hostName, err := os.Hostname()
	if err != nil {
		fmt.Printf("ERROR getting hostname for logger: %v\n", err)
	}

	return slog.New(handler.WithAttrs([]slog.Attr{
		slog.String("applicationName", appName),
		slog.String("serviceName", svcName),
		slog.String("hostName", hostName),
		slog.Int64("pid", int64(os.Getpid())),
	}))
}

func main() {
	logger := newLogger("go-slo", "job-status")

	fmt.Println(" -- initialize app")
	pgUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", userName, password, host, port, dbName)
	dbRepo, _, addCtrl, err := modinit.Init(pgUrl, logger)
	if err != nil {
		logger.Error("init failed", "err", err)
		panic(err)
	}
	defer dbRepo.Close()

	fmt.Println(" -- build mux")
	apiMux := http.NewServeMux()
	mux := http.NewServeMux()

	apiMux.Handle("/job-statuses", jshttp.Handler(logger, addCtrl))
	apiMux.Handle("/job-statuses/", jshttp.Handler(logger, addCtrl))
	mux.Handle("/api/", http.StripPrefix("/api", middleware.AddRequestId(middleware.LogRequest(apiMux, logger))))
	mux.Handle("/", logHandler(logger, "/"))

	fmt.Println(" -- start server")
	http.ListenAndServe(":9201", mux)
}
