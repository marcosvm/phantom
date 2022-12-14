package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
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
		body []byte
		err  error
	)
	body, err = io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		level.Error(h.logger).Log("msg", "error reading compressed body", "error", err.Error())
		return
	}

	origin := r.Header.Get(h.originHeader)
	origin, proxies := h.extractOrigin(origin)
	level.Debug(h.logger).Log("msg", "request received from origin", "origin", origin)

	if h.debug {
		if len(r.Header["Content-Encoding"]) > 0 && r.Header["Content-Encoding"][0] == "gzip" {
			b, err := gzip.NewReader(io.NopCloser(bytes.NewBuffer(body)))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				level.Error(h.logger).Log("msg", "error reading body", "error", err.Error())
				return
			}

			body, err = io.ReadAll(b)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				level.Error(h.logger).Log("msg", "error reading expanded body", "error", err.Error())
				return
			}
			err = b.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				level.Error(h.logger).Log("msg", "error reading expanded body", "error", err.Error())
				return
			}

		}
		var paths []struct {
			Path      string      `json:"path"`
			Value     interface{} `json:"-"`
			Timestamp interface{} `json:"-"`
		}
		err = json.Unmarshal(body, &paths)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			level.Error(h.logger).Log("msg", "error decoding JSON", "error", err.Error(), "body", string(body))
			return
		}
		for _, p := range paths {
			level.Info(h.logger).Log("msg", "metric path received", "path", p.Path)
			h.counter.With(prometheus.Labels{"origin": origin, "proxies": proxies, "path": p.Path}).Inc()
		}
	} else {
		h.counter.With(prometheus.Labels{"origin": origin, "proxies": proxies, "path": ""}).Inc()
	}
	w.WriteHeader(http.StatusOK)
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
