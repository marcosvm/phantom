package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/marcosvm/phantom/handler"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	listenAddress := flag.String("listen", ":7777", "ip:port for listening to web requests")
	originHeader := flag.String("header", "X-Forwarded-For", "request address header")
	logLevel := flag.String("log.level", "info", "debug, info, warn, error")
	debug := flag.Bool("debug", false, "print path information")
	flag.Parse()

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = level.NewFilter(logger, level.Allow(level.ParseDefault(*logLevel, level.InfoValue())))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGUSR1)
	done := make(chan struct{}, 1)

	http.Handle("/metrics", promhttp.Handler())
	h := handler.DefaultHandler(*originHeader, logger, *debug)
	http.HandleFunc("/", h.Catch)

	go func() {
		for {
			select {
			case <-sig:
				h.FlipDebug()
			case <-done:
				return
			}
		}
	}()

	level.Info(logger).Log("msg", "starting listening to requests", "address", listenAddress)
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		level.Error(logger).Log("msg", "error listening and serving", "error", err.Error())
		done <- struct{}{}
		os.Exit(1)
	}
	done <- struct{}{}
}
