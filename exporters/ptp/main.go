// PTP Exporter - Monitors PTP (IEEE 1588) clock synchronization
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PTPExporter struct {
	offsetFromMaster *prometheus.GaugeVec
	meanPathDelay    *prometheus.GaugeVec
	clockState       *prometheus.GaugeVec
	stepsRemoved     *prometheus.GaugeVec
	device           string
	interfaceName    string
}

func NewPTPExporter(device string, iface string) *PTPExporter {
	exporter := &PTPExporter{
		device:        device,
		interfaceName: iface,
		offsetFromMaster: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "st2110_ptp_offset_nanoseconds",
				Help: "Offset from PTP master clock in nanoseconds",
			},
			[]string{"device", "interface", "master"},
		),
		meanPathDelay: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "st2110_ptp_mean_path_delay_nanoseconds",
				Help: "Mean path delay to PTP master in nanoseconds",
			},
			[]string{"device", "interface", "master"},
		),
		clockState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "st2110_ptp_clock_state",
				Help: "PTP clock state (0=FREERUN, 1=LOCKED, 2=HOLDOVER)",
			},
			[]string{"device", "interface"},
		),
		stepsRemoved: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "st2110_ptp_steps_removed",
				Help: "Steps removed from grandmaster clock",
			},
			[]string{"device", "interface"},
		),
	}

	prometheus.MustRegister(exporter.offsetFromMaster)
	prometheus.MustRegister(exporter.meanPathDelay)
	prometheus.MustRegister(exporter.clockState)
	prometheus.MustRegister(exporter.stepsRemoved)

	return exporter
}

// Parse ptp4l output using pmc (PTP Management Client)
func (e *PTPExporter) CollectPTPMetrics() {
	// Execute ptp4l management query
	// Note: This requires ptp4l to be running and pmc to be installed
	cmd := exec.Command("pmc", "-u", "-b", "0", "GET", "CURRENT_DATA_SET")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to query PTP (is ptp4l running?): %v", err)
		// Set clock state to FREERUN (0) if query fails
		e.clockState.WithLabelValues(e.device, e.interfaceName).Set(0)
		return
	}

	// Parse output (example format):
	// CURRENT_DATA_SET
	//   offsetFromMaster     125
	//   meanPathDelay        523
	//   stepsRemoved         1

	outputStr := string(output)

	offsetRegex := regexp.MustCompile(`offsetFromMaster\s+(-?\d+)`)
	delayRegex := regexp.MustCompile(`meanPathDelay\s+(\d+)`)
	stepsRegex := regexp.MustCompile(`stepsRemoved\s+(\d+)`)
	stateRegex := regexp.MustCompile(`clockState\s+(\w+)`)

	if matches := offsetRegex.FindStringSubmatch(outputStr); len(matches) > 1 {
		if offset, err := strconv.ParseFloat(matches[1], 64); err == nil {
			e.offsetFromMaster.WithLabelValues(e.device, e.interfaceName, "grandmaster").Set(offset)
		}
	}

	if matches := delayRegex.FindStringSubmatch(outputStr); len(matches) > 1 {
		if delay, err := strconv.ParseFloat(matches[1], 64); err == nil {
			e.meanPathDelay.WithLabelValues(e.device, e.interfaceName, "grandmaster").Set(delay)
		}
	}

	if matches := stepsRegex.FindStringSubmatch(outputStr); len(matches) > 1 {
		if steps, err := strconv.ParseFloat(matches[1], 64); err == nil {
			e.stepsRemoved.WithLabelValues(e.device, e.interfaceName).Set(steps)
		}
	}

	// Parse clock state
	if matches := stateRegex.FindStringSubmatch(outputStr); len(matches) > 1 {
		state := matches[1]
		var stateValue float64
		switch state {
		case "LOCKED":
			stateValue = 1
		case "HOLDOVER":
			stateValue = 2
		default:
			stateValue = 0 // FREERUN
		}
		e.clockState.WithLabelValues(e.device, e.interfaceName).Set(stateValue)
	} else {
		// Default to LOCKED if state not found but query succeeded
		e.clockState.WithLabelValues(e.device, e.interfaceName).Set(1)
	}
}

func (e *PTPExporter) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		// Collect immediately
		e.CollectPTPMetrics()
		// Then collect at intervals
		for range ticker.C {
			e.CollectPTPMetrics()
		}
	}()
}

func main() {
	device := flag.String("device", "default", "Device name identifier")
	iface := flag.String("interface", "eth0", "Network interface name")
	listenAddr := flag.String("listen", ":9200", "Prometheus exporter listen address")
	interval := flag.Duration("interval", 1*time.Second, "PTP metrics collection interval")
	flag.Parse()

	// Allow override from environment
	if envDevice := os.Getenv("DEVICE"); envDevice != "" {
		device = &envDevice
	}
	if envInterface := os.Getenv("INTERFACE"); envInterface != "" {
		iface = &envInterface
	}
	if envListen := os.Getenv("LISTEN_ADDR"); envListen != "" {
		listenAddr = &envListen
	}

	exporter := NewPTPExporter(*device, *iface)
	exporter.Start(*interval)

	log.Printf("Starting PTP exporter on %s (device: %s, interface: %s)", *listenAddr, *device, *iface)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK\n")
	})

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
