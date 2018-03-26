package exporter

import (
	"github.com/blockassets/cgminer_client"
	"github.com/blockassets/prometheus_helper"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"reflect"
	"strings"
	"sync"
)

//
var (
	namespace    = "cgminer"
	idLabelNames = []string{"id"}
)

// Collector interface
type Exporter struct {
	client      cgminer_client.Client
	ConstLabels prometheus.Labels
	Gauges      prometheus_helper.GaugeMapMap
	GaugeVecs   prometheus_helper.GaugeVecMapMap
	ChipStats   *prometheus.GaugeVec
	sync.Mutex
}

type MinerData struct {
	Summary cgminer_client.Summary
	Devs    map[string]cgminer_client.Dev
}

//
func NewExporter(client cgminer_client.Client, version string) *Exporter {
	constLabels := prometheus.Labels{"version": version}

	structFieldMap := prometheus_helper.NewStructFieldMap(MinerData{})

	chipStats := prometheus_helper.NewGaugeVec(
		namespace,
		"chipstats_accept",
		"ChipStats Accept",
		constLabels, []string{"id", "chip"})

	return &Exporter{
		client:      client,
		ConstLabels: constLabels,
		Gauges:      prometheus_helper.NewGaugeMapMap(structFieldMap, namespace, constLabels),
		GaugeVecs:   make(prometheus_helper.GaugeVecMapMap),
		ChipStats:   &chipStats,
	}
}

//
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	prometheus_helper.DescribeGaugeMapMap(e.Gauges, ch)
	prometheus_helper.DescribeGaugeVecMapMap(e.GaugeVecs, ch)
	e.ChipStats.Describe(ch)
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

	minerData := &MinerData{
		Summary: *summary,
		Devs:    mapDevs,
	}

	minerDataMap := prometheus_helper.NewStructFieldMap(*minerData)

	for key, value := range minerDataMap {
		val := reflect.ValueOf(value)
		// 'Devs' is a special case as a GaugeVec
		if val.Kind() == reflect.Map {
			for _, k := range val.MapKeys() {
				worker := val.MapIndex(k).Interface()
				labelValues := prometheus.Labels{idLabelNames[0]: k.Interface().(string)}
				prometheus_helper.CollectGaugeVecs(key, worker, e.GaugeVecs, namespace, e.ConstLabels, idLabelNames, labelValues)
			}
		} else {
			meta := prometheus_helper.StructMeta{}
			prometheus_helper.MakeStructMeta(value, &meta)
			prometheus_helper.SetValuesOnGauges(meta, namespace, e.Gauges[key])
		}
	}

	for _, chipStat := range *chipStats {
		for chipName, chipValue := range chipStat.Accept {
			chipId := strings.Split(chipName, "_")[0]
			flt, err := prometheus_helper.ConvertToFloat(chipValue)
			if err == nil {
				e.ChipStats.WithLabelValues(chipStat.Name, chipId).Set(flt)
			}
		}
	}

	e.ChipStats.Collect(ch)

	prometheus_helper.CollectGaugeMapMap(e.Gauges, ch)
	prometheus_helper.CollectGaugeVecMapMap(e.GaugeVecs, ch)
}
