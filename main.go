package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/blockassets/cgminer_client"
	"github.com/blockassets/cgminer_exporter/exporter"
)

var (
	// Makefile build
	version = ""
)

func main() {
	port := flag.String("port", "4030", "The address to listen on for /metrics HTTP requests.")
	cgHost := flag.String("cghost", "127.0.0.1", "The address of the worker.")
	cgPort := flag.Int64("cgport", 4028, "The port cgminer runs on.")
	cgTimeout := flag.Duration("cgtimeout", 5*time.Second, "The amount of time to wait for cgminer to return.")
	flag.Parse()

	cgVersion := ReadVersionFile()
	if len(cgVersion) == 0 {
		cgVersion = "unknown"
	}

	client := cgminer_client.New(*cgHost, *cgPort, *cgTimeout)
	exporter := exporter.NewExporter(client, cgVersion)

	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("%s %s", os.Args[0], version)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *port), nil))
}

//
func readFileTrim(file string) string {
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println(err)
	}

	return strings.TrimSpace(string(dat))
}

/*
	BW saves their cgminer version into a file.
*/
func ReadVersionFile() string {
	return readFileTrim("/usr/app/version.txt")
}
