package evertz

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type EvertzEQXExporter struct {
	target     string
	snmp       *gosnmp.GoSNMP
	httpClient *http.Client

	// Evertz-specific metrics
	moduleStatus      *prometheus.GaugeVec
	ipFlowStatus      *prometheus.GaugeVec
	videoStreamStatus *prometheus.GaugeVec
	ptpStatus         *prometheus.GaugeVec
	redundancyStatus  *prometheus.GaugeVec
}

func NewEvertzEQXExporter(target, snmpCommunity string) *EvertzEQXExporter {
	snmp := &gosnmp.GoSNMP{
		Target:    target,
		Port:      161,
		Community: snmpCommunity,
		Version:   gosnmp.Version2c,
		Timeout:   5 * time.Second,
	}

	return &EvertzEQXExporter{
		target:     target,
		snmp:       snmp,
		httpClient: &http.Client{Timeout: 10 * time.Second},

		moduleStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "evertz_eqx_module_status",
				Help: "EQX module status (1=OK, 0=fault)",
			},
			[]string{"chassis", "slot", "module_type"},
		),

		ipFlowStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "evertz_eqx_ip_flow_status",
				Help: "IP flow status (1=active, 0=inactive)",
			},
			[]string{"chassis", "flow_id", "direction"},
		),

		videoStreamStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "evertz_eqx_video_stream_status",
				Help: "Video stream status (1=present, 0=no signal)",
			},
			[]string{"chassis", "stream_id"},
		),

		ptpStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "evertz_eqx_ptp_lock_status",
				Help: "PTP lock status (1=locked, 0=unlocked)",
			},
			[]string{"chassis", "module"},
		),

		redundancyStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "evertz_eqx_redundancy_status",
				Help: "Redundancy status (1=protected, 0=unprotected)",
			},
			[]string{"chassis", "pair"},
		),
	}
}

// Evertz EQX uses both SNMP and HTTP XML API
func (e *EvertzEQXExporter) Collect() error {
	// Connect SNMP
	if err := e.snmp.Connect(); err != nil {
		return err
	}
	defer e.snmp.Conn.Close()

	// Walk Evertz MIB tree
	if err := e.collectSNMP(); err != nil {
		return err
	}

	// Get detailed status via HTTP XML API
	if err := e.collectHTTPAPI(); err != nil {
		return err
	}

	return nil
}

// Evertz-specific SNMP OIDs
const (
	evertzModuleStatusOID = ".1.3.6.1.4.1.6827.20.1.1.1.1.2" // evModule Status
	evertzIPFlowStatusOID = ".1.3.6.1.4.1.6827.20.2.1.1.1.5" // evIPFlow Status
	evertzPTPLockOID      = ".1.3.6.1.4.1.6827.20.3.1.1.1.3" // evPTP Lock Status
)

func (e *EvertzEQXExporter) collectSNMP() error {
	// Walk module status
	err := e.snmp.Walk(evertzModuleStatusOID, func(pdu gosnmp.SnmpPDU) error {
		// Parse OID to extract chassis/slot
		chassis, slot := parseEvertzOID(pdu.Name)
		status := pdu.Value.(int)

		e.moduleStatus.WithLabelValues(chassis, slot, "unknown").Set(float64(status))
		return nil
	})

	return err
}

func (e *EvertzEQXExporter) collectHTTPAPI() error {
	// Evertz XML API endpoint
	url := fmt.Sprintf("http://%s/status.xml", e.target)

	resp, err := e.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var status EvertzStatus
	if err := xml.NewDecoder(resp.Body).Decode(&status); err != nil {
		return err
	}

	// Update Prometheus metrics from XML
	for _, flow := range status.IPFlows {
		e.ipFlowStatus.WithLabelValues(
			status.Chassis,
			flow.ID,
			flow.Direction,
		).Set(boolToFloat(flow.Active))
	}

	return nil
}

type EvertzStatus struct {
	Chassis string   `xml:"chassis,attr"`
	IPFlows []IPFlow `xml:"ipflows>flow"`
}

type IPFlow struct {
	ID        string `xml:"id,attr"`
	Direction string `xml:"direction,attr"`
	Active    bool   `xml:"active"`
}

func parseEvertzOID(oid string) (chassis, slot string) {
	// Parse Evertz OID format
	// Example: .1.3.6.1.4.1.6827.20.1.1.1.1.2.1.5 -> chassis 1, slot 5
	parts := splitOID(oid)
	if len(parts) >= 14 {
		return parts[13], parts[14]
	}
	return "1", "5" // Simplified
}

func splitOID(oid string) []string {
	// Simple OID parsing - implement proper parsing
	return []string{}
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func main() {
	target := flag.String("target", "", "Evertz EQX target IP")
	community := flag.String("community", "public", "SNMP community")
	listenAddr := flag.String("listen", ":9500", "Prometheus exporter listen address")
	flag.Parse()

	exporter := NewEvertzEQXExporter(*target, *community)

	// Register metrics
	prometheus.MustRegister(exporter.moduleStatus)
	prometheus.MustRegister(exporter.ipFlowStatus)
	prometheus.MustRegister(exporter.videoStreamStatus)
	prometheus.MustRegister(exporter.ptpStatus)
	prometheus.MustRegister(exporter.redundancyStatus)

	// Collect metrics periodically
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			if err := exporter.Collect(); err != nil {
				log.Printf("Collection error: %v", err)
			}
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting Evertz EQX exporter on %s", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
