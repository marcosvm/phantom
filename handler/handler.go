package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Handler contains the origin string configuration, logger and Prometheus counter
type Handler struct {
	originHeader string
	logger       log.Logger
	counter      *prometheus.CounterVec
	debug        bool
}

// DefaultHandler returns a *Handler with default configuration
func DefaultHandler(originHeader string, logger log.Logger, debug bool) *Handler {
	return &Handler{
		originHeader: originHeader,
		logger:       logger,
		counter: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "metrics_posts_received_total",
			Help: "The total number of received posts for metrics",
		}, []string{
			"origin",
			"proxies",
			"path",
		}),
		debug: debug,
	}
}

// NewHandler returns a *Handler with custom parameters, useful for testing
func NewHandler(originHeader string, logger log.Logger, counter *prometheus.CounterVec, debug bool) *Handler {
	return &Handler{
		originHeader: originHeader,
		logger:       logger,
		counter:      counter,
		debug:        debug,
	}
}

// Catch handles all requests and increments the Prometheus counter with an origin label
func (h *Handler) Catch(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r.Body != nil {
			r.Body.Close()
		}
	}()
	var (
		err error
	)
	_, err = io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		level.Error(h.logger).Log("msg", "error reading compressed body", "error", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
   "took": 30,
   "errors": false,
   "items": [
      {
         "index": {
            "_index": "test",
            "_id": "1",
            "_version": 1,
            "result": "created",
            "_shards": {
               "total": 2,
               "successful": 1,
               "failed": 0
            },
            "status": 201,
            "_seq_no" : 0,
            "_primary_term": 1
         }
      }]}
`)

}

func (h *Handler) extractOrigin(origin string) (string, string) {
	if origin == "" {
		return "unknown", ""
	}

	parts := strings.SplitN(origin, ",", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return strings.Trim(parts[0], " "), strings.Trim(parts[1], " ")
}

func (h *Handler) FlipDebug() {
	h.debug = !h.debug
}
