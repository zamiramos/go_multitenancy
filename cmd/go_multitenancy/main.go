package main

import (
	"flag"
	"fmt"
	"go_multitenancy/car"
	"go_multitenancy/db"
	"go_multitenancy/middlewares"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

import (
	"github.com/go-kit/kit/log"
)

func main() {
	var (
		httpAddr = flag.String("http.addr", "127.0.0.1:8080", "HTTP listen address")
		_        = flag.String("db.sqlite.path", ".\\", "Path to SQLite db file")
		dbType   = flag.String("db.type", "sqlite", "Type of Database: sqlite, postgresql")
	)
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var dbHandler db.Handler
	{
		var err error
		dbHandler, err = db.HandlerFactory(*dbType)
		if err != nil {
			logger.Log("exit", err)
			os.Exit(-1)
		}
	}

	var s car.Service
	{
		repo := car.NewRepo(dbHandler, logger)

		s = car.NewInmemService(repo)
		s = car.LoggingMiddleware(logger)(s)
	}

	var h http.Handler
	{
		logger := log.With(logger, "component", "HTTP")
		h = car.MakeHTTPHandler(s, logger, middlewares.CreateTenantResolverMiddleware(logger))
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errs <- http.ListenAndServe(*httpAddr, h)
	}()

	logger.Log("exit", <-errs)
}
