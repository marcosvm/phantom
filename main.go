package main

import (
	"flag"
	"net/http"
	"os"

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

	http.Handle("/metrics", promhttp.Handler())

	handler := handler.DefaultHandler(*originHeader, logger, *debug).Catch
	http.HandleFunc("/", handler)

	level.Info(logger).Log("msg", "starting listening to requests", "address", listenAddress)
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		level.Error(logger).Log("msg", "error listening and serving", "error", err.Error())
		os.Exit(1)
	}
}
