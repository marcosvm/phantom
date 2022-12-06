package handler

import (
	"encoding/json"
	"io/ioutil"
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
			"proxies",
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
	buf, err := ioutil.ReadAll(r.Body)
	var paths []struct {
		Path      string      `json:"path"`
		Value     interface{} `json:"-"`
		Timestamp interface{} `json:"-"`
	}
	err = json.Unmarshal(buf, &paths)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, p := range paths {
		level.Info(h.logger).Log("msg", "metric path received", "path", p.Path)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	origin := r.Header.Get(h.originHeader)
	origin, proxies := h.extractOrigin(origin)
	level.Debug(h.logger).Log("msg", "request received from origin", "origin", origin)
	h.counter.With(prometheus.Labels{"origin": origin, "proxies": proxies}).Inc()
	w.WriteHeader(http.StatusOK)
}

func (h Handler) extractOrigin(origin string) (string, string) {
	if origin == "" {
		return "unknown", ""
	}

	parts := strings.SplitN(origin, ",", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return strings.Trim(parts[0], " "), strings.Trim(parts[1], " ")
}
