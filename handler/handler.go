package handler

import (
	"io/ioutil"
	"net/http"

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
}

// DefaultHandler returns a *Handler with default configuration
func DefaultHandler(originHeader string, logger log.Logger) *Handler {
	return &Handler{
		originHeader: originHeader,
		logger:       logger,
		counter: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "metrics_posts_received_total",
			Help: "The total number of received posts for metrics",
		}, []string{
			"origin",
		}),
	}
}

// NewHandler returns a *Handler with custom parameters, useful for testing
func NewHandler(originHeader string, logger log.Logger, counter *prometheus.CounterVec) *Handler {
	return &Handler{
		originHeader: originHeader,
		logger:       logger,
		counter:      counter,
	}
}

// Catch handles all requests and increments the Prometheus counter with an origin label
func (h Handler) Catch(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r.Body != nil {
			r.Body.Close()
		}
	}()
	_, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	origin := r.Header.Get(h.originHeader)
	if origin != "" {
		level.Debug(h.logger).Log("msg", "request received from origin", "origin", origin)
		h.counter.With(prometheus.Labels{"origin": origin}).Inc()
	}
	w.WriteHeader(http.StatusOK)
}
