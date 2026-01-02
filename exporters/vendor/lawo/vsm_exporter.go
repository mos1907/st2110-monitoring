package lawo

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type LawoVSMExporter struct {
	baseURL  string // http://vsm-server:9000
	apiToken string

	// VSM-specific metrics
	connectionStatus *prometheus.GaugeVec
	deviceStatus     *prometheus.GaugeVec
	pathwayStatus    *prometheus.GaugeVec
	alarmCount       *prometheus.GaugeVec
}

func NewLawoVSMExporter(baseURL, apiToken string) *LawoVSMExporter {
	return &LawoVSMExporter{
		baseURL:  baseURL,
		apiToken: apiToken,

		connectionStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "lawo_vsm_connection_status",
				Help: "VSM connection status (1=connected, 0=disconnected)",
			},
			[]string{"device_name", "device_type"},
		),

		deviceStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "lawo_vsm_device_status",
				Help: "Device status (1=OK, 0=fault)",
			},
			[]string{"device_name", "device_type"},
		),

		pathwayStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "lawo_vsm_pathway_status",
				Help: "Signal pathway status (1=active, 0=inactive)",
			},
			[]string{"pathway_name", "source", "destination"},
		),

		alarmCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "lawo_vsm_active_alarms",
				Help: "Number of active alarms",
			},
			[]string{"severity"},
		),
	}
}

func (e *LawoVSMExporter) Collect() error {
	// Get device tree from VSM
	devices, err := e.getDevices()
	if err != nil {
		return err
	}

	for _, device := range devices {
		e.deviceStatus.WithLabelValues(device.Name, device.Type).Set(
			boolToFloat(device.Status == "OK"),
		)
		e.connectionStatus.WithLabelValues(device.Name, device.Type).Set(
			boolToFloat(device.Connected),
		)
	}

	// Get active pathways
	pathways, err := e.getPathways()
	if err != nil {
		return err
	}

	for _, pathway := range pathways {
		e.pathwayStatus.WithLabelValues(
			pathway.Name,
			pathway.Source,
			pathway.Destination,
		).Set(boolToFloat(pathway.Active))
	}

	// Get alarm summary
	alarms, err := e.getAlarms()
	if err != nil {
		return err
	}

	alarmCounts := map[string]int{"critical": 0, "warning": 0, "info": 0}
	for _, alarm := range alarms {
		alarmCounts[alarm.Severity]++
	}

	for severity, count := range alarmCounts {
		e.alarmCount.WithLabelValues(severity).Set(float64(count))
	}

	return nil
}

// VSM REST API client
func (e *LawoVSMExporter) makeRequest(endpoint string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v1/%s", e.baseURL, endpoint)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+e.apiToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body []byte
	resp.Body.Read(body)

	return body, nil
}

func (e *LawoVSMExporter) getDevices() ([]VSMDevice, error) {
	data, err := e.makeRequest("devices")
	if err != nil {
		return nil, err
	}

	var result struct {
		Devices []VSMDevice `json:"devices"`
	}
	json.Unmarshal(data, &result)

	return result.Devices, nil
}

func (e *LawoVSMExporter) getPathways() ([]VSMPathway, error) {
	data, err := e.makeRequest("pathways")
	if err != nil {
		return nil, err
	}

	var result struct {
		Pathways []VSMPathway `json:"pathways"`
	}
	json.Unmarshal(data, &result)

	return result.Pathways, nil
}

func (e *LawoVSMExporter) getAlarms() ([]VSMAlarm, error) {
	data, err := e.makeRequest("alarms?state=active")
	if err != nil {
		return nil, err
	}

	var result struct {
		Alarms []VSMAlarm `json:"alarms"`
	}
	json.Unmarshal(data, &result)

	return result.Alarms, nil
}

type VSMDevice struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Connected bool   `json:"connected"`
}

type VSMPathway struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Active      bool   `json:"active"`
}

type VSMAlarm struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Device   string `json:"device"`
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func main() {
	baseURL := flag.String("url", "http://vsm-server:9000", "VSM base URL")
	apiToken := flag.String("token", "", "VSM API token")
	listenAddr := flag.String("listen", ":9600", "Prometheus exporter listen address")
	flag.Parse()

	exporter := NewLawoVSMExporter(*baseURL, *apiToken)

	// Register metrics
	prometheus.MustRegister(exporter.connectionStatus)
	prometheus.MustRegister(exporter.deviceStatus)
	prometheus.MustRegister(exporter.pathwayStatus)
	prometheus.MustRegister(exporter.alarmCount)

	// Collect metrics periodically
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		for range ticker.C {
			if err := exporter.Collect(); err != nil {
				log.Printf("Collection error: %v", err)
			}
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting Lawo VSM exporter on %s", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
