package main

import (
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"go_multitenancy/car"
	"go_multitenancy/db"
	"go_multitenancy/middlewares"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

import (
	"github.com/go-kit/kit/log"
)

func main() {
	var (
		httpAddr       = flag.String("http.addr", "127.0.0.1:8080", "HTTP listen address")
		_              = flag.String("db.sqlite.path", ".\\", "Path to SQLite db file")
		dbType         = flag.String("db.type", "sqlite", "Type of Database: sqlite, postgresql")
		mode           = flag.String("mode", "start", "supported mode are kill and start")
		pid            = flag.Int("pid", 0, "the pid of he process")
		exit_sync_file = ".\\ciam_int"
	)
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	if mode != nil && *mode == "kill" {
		if pid == nil || *pid == 0 {
			logger.Log("error", "pid must be specified ")
			os.Exit(-1)
		}

		proc, err := os.FindProcess(*pid)
		if err != nil {
			logger.Log(err)
			os.Exit(-1)
		}

		os.Remove(exit_sync_file)

		time.AfterFunc(3*time.Second, func() {
			proc.Kill()
		})

		proc.Wait()

		// Kill the process
		os.Exit(0)
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
		// Create new watcher.
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			errs <- fmt.Errorf("error %s", err)
		}
		defer watcher.Close()

		// create file if not exit
		f, err := os.OpenFile(exit_sync_file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			errs <- fmt.Errorf("error %s", err)
		}

		f.Close()

		err = watcher.Add(exit_sync_file)
		if err != nil {
			errs <- fmt.Errorf("error %s", err)
		}

		for {
			event, ok := <-watcher.Events
			if !ok {
				errs <- fmt.Errorf("failed listen to interpret file")
				break
			}

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Remove) {
				break
			}
		}

		errs <- fmt.Errorf("process intturpted")
	}()

	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errs <- http.ListenAndServe(*httpAddr, h)
	}()

	logger.Log("exit", <-errs)
}
