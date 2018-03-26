package exporter

import (
	"fmt"
	"github.com/blockassets/cgminer_client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewExporter(t *testing.T) {
	cgClient := cgminer_client.New("10.10.0.11", 4028, time.Duration(2)*time.Second)
	exporter := NewExporter(cgClient, "1.0")

	reg := prometheus.NewRegistry()
	reg.MustRegister(exporter)

	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()

	body, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)

	fmt.Println(bodyStr)
}
