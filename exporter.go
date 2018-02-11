package main

import (
	"log"
	"sync"
	"time"

	"github.com/blockassets/cgminer_client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/blockassets/prometheus_helper"
	"reflect"
)

//
var (
	namespace    = "cgminer"
	idLabelNames = []string{"id"}
)

// Collector interface
type Exporter struct {
	client      *cgminer_client.Client
	ConstLabels prometheus.Labels
	Gauges      prometheus_helper.GaugeMapMap
	GaugeVecs   prometheus_helper.GaugeVecMapMap
	sync.Mutex
}

type MinerData struct {
	Summary cgminer_client.Summary
	Devs map[string]cgminer_client.Dev
	ChipStats map[string]cgminer_client.ChipStat
}

//
func NewExporter(host string, port int64, timeout time.Duration, version string) *Exporter {
	constLabels := prometheus.Labels{"version": version}

	structFieldMap := prometheus_helper.NewStructFieldMap(MinerData{})

	return &Exporter{
		client:      cgminer_client.New(host, port, timeout),
		ConstLabels: constLabels,
		Gauges:      prometheus_helper.NewGaugeMapMap(structFieldMap, namespace, constLabels),
		GaugeVecs:   make(prometheus_helper.GaugeVecMapMap),
	}
}

//
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	prometheus_helper.DescribeGaugeMapMap(e.Gauges, ch)
	prometheus_helper.DescribeGaugeVecMapMap(e.GaugeVecs, ch)
}

//
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	// Prevents multiple concurrent calls
	e.Lock()
	defer e.Unlock()

	summary, err := e.client.Summary()
	if err != nil {
		log.Println(err)
		return
	}

	devs, err := e.client.Devs()
	if err != nil {
		log.Println(err)
		return
	}

	chipStats, err := e.client.ChipStat()
	if err != nil {
		log.Println(err)
		return
	}

	mapDevs := make(map[string]cgminer_client.Dev)
	for _, val := range *devs {
		mapDevs[val.Name] = val
	}

	mapChipStats := make(map[string]cgminer_client.ChipStat)
	for _, val := range *chipStats {
		mapChipStats[val.Name] = val
	}

	minerData := &MinerData{
		Summary:   *summary,
		Devs:      mapDevs,
		ChipStats: mapChipStats,
	}

	poolDataMap := prometheus_helper.NewStructFieldMap(*minerData)

	for key, value := range poolDataMap {
		val := reflect.ValueOf(value)
		// 'Devs' and 'ChipStats' is a special case as a GaugeVec
		if val.Kind() == reflect.Map {
			for _, key := range val.MapKeys() {
				name := key.Interface().(string)
				worker := val.MapIndex(key).Interface()
				labelValues := prometheus.Labels{idLabelNames[0]: name}
				prometheus_helper.CollectGaugeVecs(name, worker, e.GaugeVecs, namespace, e.ConstLabels, idLabelNames, labelValues)
			}
		} else {
			meta := prometheus_helper.StructMeta{}
			prometheus_helper.MakeStructMeta(value, &meta)
			prometheus_helper.SetValuesOnGauges(meta, namespace, e.Gauges[key])
		}
	}

	prometheus_helper.CollectGaugeMapMap(e.Gauges, ch)
	prometheus_helper.CollectGaugeVecMapMap(e.GaugeVecs, ch)
}
