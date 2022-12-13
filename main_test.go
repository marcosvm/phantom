package main_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/marcosvm/phantom/handler"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func TestHandler(t *testing.T) {
	jsonBody := `
	{
		"path": "hosts.metric.uno",
		"value": "33.441",
		"timestamp": "1609746000"
	},
	{
		"path": "hosts.metric.uno",
		"value": "41",
		"timestamp": "1609786088"
	},
	{
		"path": "hosts.metric.uno",
		"value": "11341341414.3",
		"timestamp": "1609746210"
	}
	`

	r := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v2/graphite", strings.NewReader(jsonBody))
	req.Header.Add("X-Forwarded-For", "10.1.1.1")

	handler := handler.DefaultHandler("X-Forwarded-For", log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr)))
	handler.Catch(r, req)

	if r.Result().StatusCode != 200 {
		t.Errorf("failed to process request, expected 200 but got %d", r.Result().StatusCode)
	}
}

func TestEmptyBody(t *testing.T) {
	r := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v2/graphite", nil)

	handler := handler.NewHandler(
		"X-Forwarded-For",
		log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr)),
		promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "metrics_posts_received_total_test_edition",
			Help: "The total number of received posts for metrics",
		}, []string{
			"origin",
			"proxies",
		}),
	)

	handler.Catch(r, req)

	if r.Result().StatusCode != 200 {
		t.Errorf("failed to process request, expected 200 but got %d", r.Result().StatusCode)
	}
}

var (
	benchLogger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	errorLogger = level.NewFilter(benchLogger, level.AllowError())

	benchHandler = handler.NewHandler(
		"X-Forwarded-For",
		errorLogger,
		promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "metrics_posts_received_total_benchmark_edition",
			Help: "The total number of received posts for metrics",
		}, []string{
			"origin",
			"proxies",
		}),
	)
)

// go test -bench -cpu=1,2,4,8,16,32 -run=^$ -bench ^BenchmarkPost$ github.com/marcosvm/phantom
func BenchmarkPost(b *testing.B) {
	jsonBody := `
	{
		"path": "hosts.metric.uno",
		"value": "33.441",
		"timestamp": "1609746000"
	},
	{
		"path": "hosts.metric.uno",
		"value": "41",
		"timestamp": "1609786088"
	},
	{
		"path": "hosts.metric.uno",
		"value": "11341341414.3",
		"timestamp": "1609746210"
	}
        `

	b.RunParallel(func(pb *testing.PB) {
		r, _ := http.NewRequest("POST", "/api/v2/graphite", strings.NewReader(jsonBody))
		r.Header.Add("X-Forwarded-For", "10.1.1.1")

		w := httptest.NewRecorder()
		h := http.HandlerFunc(benchHandler.Catch)

		b.ReportAllocs()
		b.ResetTimer()

		for pb.Next() {
			h.ServeHTTP(w, r)
		}
	})
}
