package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/yaml.v2"
)

type SwitchConfig struct {
	Name     string `yaml:"name"`
	Target   string `yaml:"target"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Vendor   string `yaml:"vendor"`
}

type Config struct {
	Switches []SwitchConfig `yaml:"switches"`
}

type GNMICollector struct {
	target   string
	username string
	password string

	// Prometheus metrics
	interfaceRxBytes  *prometheus.GaugeVec
	interfaceTxBytes  *prometheus.GaugeVec
	interfaceRxErrors *prometheus.GaugeVec
	interfaceTxErrors *prometheus.GaugeVec
	interfaceRxDrops  *prometheus.GaugeVec
	interfaceTxDrops  *prometheus.GaugeVec
	qosBufferUtil     *prometheus.GaugeVec
	qosDroppedPackets *prometheus.GaugeVec
	multicastGroups   *prometheus.GaugeVec
}

func NewGNMICollector(target, username, password string) *GNMICollector {
	collector := &GNMICollector{
		target:   target,
		username: username,
		password: password,

		interfaceRxBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "st2110_switch_interface_rx_bytes",
				Help: "Received bytes on switch interface",
			},
			[]string{"switch", "interface"},
		),

		interfaceTxBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "st2110_switch_interface_tx_bytes",
				Help: "Transmitted bytes on switch interface",
			},
			[]string{"switch", "interface"},
		),

		interfaceRxErrors: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "st2110_switch_interface_rx_errors",
				Help: "Receive errors on switch interface",
			},
			[]string{"switch", "interface"},
		),

		interfaceTxErrors: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "st2110_switch_interface_tx_errors",
				Help: "Transmit errors on switch interface",
			},
			[]string{"switch", "interface"},
		),

		interfaceRxDrops: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "st2110_switch_interface_rx_drops",
				Help: "Dropped received packets on switch interface",
			},
			[]string{"switch", "interface"},
		),

		interfaceTxDrops: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "st2110_switch_interface_tx_drops",
				Help: "Dropped transmitted packets on switch interface",
			},
			[]string{"switch", "interface"},
		),

		qosBufferUtil: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "st2110_switch_qos_buffer_utilization",
				Help: "QoS buffer utilization percentage",
			},
			[]string{"switch", "interface", "queue"},
		),

		qosDroppedPackets: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "st2110_switch_qos_dropped_packets",
				Help: "Packets dropped due to QoS",
			},
			[]string{"switch", "interface", "queue"},
		),

		multicastGroups: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "st2110_switch_multicast_groups",
				Help: "Number of IGMP multicast groups",
			},
			[]string{"switch", "interface"},
		),
	}

	// Register metrics
	prometheus.MustRegister(collector.interfaceRxBytes)
	prometheus.MustRegister(collector.interfaceTxBytes)
	prometheus.MustRegister(collector.interfaceRxErrors)
	prometheus.MustRegister(collector.interfaceTxErrors)
	prometheus.MustRegister(collector.interfaceRxDrops)
	prometheus.MustRegister(collector.interfaceTxDrops)
	prometheus.MustRegister(collector.qosBufferUtil)
	prometheus.MustRegister(collector.qosDroppedPackets)
	prometheus.MustRegister(collector.multicastGroups)

	return collector
}

func (c *GNMICollector) Connect() (gnmi.GNMIClient, error) {
	// TLS configuration (skip verification for lab, use proper certs in production!)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // ⚠️ Use proper certificates in production
	}

	// gRPC connection options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		grpc.WithPerRPCCredentials(&loginCreds{
			Username: c.username,
			Password: c.password,
		}),
		grpc.WithBlock(),
		grpc.WithTimeout(10 * time.Second),
	}

	// Connect to gNMI target
	conn, err := grpc.Dial(c.target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", c.target, err)
	}

	client := gnmi.NewGNMIClient(conn)
	log.Printf("Connected to gNMI target: %s", c.target)

	return client, nil
}

func (c *GNMICollector) Subscribe(ctx context.Context) error {
	client, err := c.Connect()
	if err != nil {
		return err
	}

	// Create subscription request
	subscribeReq := &gnmi.SubscribeRequest{
		Request: &gnmi.SubscribeRequest_Subscribe{
			Subscribe: &gnmi.SubscriptionList{
				Mode: gnmi.SubscriptionList_STREAM,
				Subscription: []*gnmi.Subscription{
					// Interface counters
					{
						Path: &gnmi.Path{
							Elem: []*gnmi.PathElem{
								{Name: "interfaces"},
								{Name: "interface", Key: map[string]string{"name": "*"}},
								{Name: "state"},
								{Name: "counters"},
							},
						},
						Mode:           gnmi.SubscriptionMode_SAMPLE,
						SampleInterval: 1000000000, // 1 second in nanoseconds
					},
					// QoS queue statistics
					{
						Path: &gnmi.Path{
							Elem: []*gnmi.PathElem{
								{Name: "qos"},
								{Name: "interfaces"},
								{Name: "interface", Key: map[string]string{"name": "*"}},
								{Name: "output"},
								{Name: "queues"},
								{Name: "queue", Key: map[string]string{"name": "*"}},
								{Name: "state"},
							},
						},
						Mode:           gnmi.SubscriptionMode_SAMPLE,
						SampleInterval: 1000000000, // 1 second
					},
				},
				Encoding: gnmi.Encoding_JSON_IETF,
			},
		},
	}

	// Start subscription stream
	stream, err := client.Subscribe(ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Send subscription request
	if err := stream.Send(subscribeReq); err != nil {
		return fmt.Errorf("failed to send subscription: %w", err)
	}

	log.Println("Started gNMI subscription stream")

	// Receive updates
	for {
		response, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("stream error: %w", err)
		}

		c.handleUpdate(response)
	}
}

func (c *GNMICollector) handleUpdate(response *gnmi.SubscribeResponse) {
	switch resp := response.Response.(type) {
	case *gnmi.SubscribeResponse_Update:
		notification := resp.Update

		// Extract switch name from prefix
		switchName := c.target

		for _, update := range notification.Update {
			path := update.Path
			value := update.Val

			// Parse interface counters
			if len(path.Elem) >= 4 && path.Elem[0].Name == "interfaces" {
				ifaceName := path.Elem[1].Key["name"]

				if path.Elem[2].Name == "state" && path.Elem[3].Name == "counters" {
					// Parse counter values from JSON
					if jsonVal := value.GetJsonIetfVal(); jsonVal != nil {
						counters := parseCounters(jsonVal)
						c.interfaceRxBytes.WithLabelValues(switchName, ifaceName).Set(float64(counters.InOctets))
						c.interfaceTxBytes.WithLabelValues(switchName, ifaceName).Set(float64(counters.OutOctets))
						c.interfaceRxErrors.WithLabelValues(switchName, ifaceName).Set(float64(counters.InErrors))
						c.interfaceTxErrors.WithLabelValues(switchName, ifaceName).Set(float64(counters.OutErrors))
						c.interfaceRxDrops.WithLabelValues(switchName, ifaceName).Set(float64(counters.InDiscards))
						c.interfaceTxDrops.WithLabelValues(switchName, ifaceName).Set(float64(counters.OutDiscards))
					}
				}
			}

			// Parse QoS queue statistics
			if len(path.Elem) >= 7 && path.Elem[0].Name == "qos" {
				ifaceName := path.Elem[2].Key["name"]
				queueName := path.Elem[5].Key["name"]

				if jsonVal := value.GetJsonIetfVal(); jsonVal != nil {
					qos := parseQoSStats(jsonVal)
					c.qosBufferUtil.WithLabelValues(switchName, ifaceName, queueName).Set(qos.BufferUtilization)
					c.qosDroppedPackets.WithLabelValues(switchName, ifaceName, queueName).Set(float64(qos.DroppedPackets))
				}
			}
		}

	case *gnmi.SubscribeResponse_SyncResponse:
		log.Println("Received sync response (initial sync complete)")
	}
}

// Helper structures
type InterfaceCounters struct {
	InOctets    uint64
	OutOctets   uint64
	InErrors    uint64
	OutErrors   uint64
	InDiscards  uint64
	OutDiscards uint64
}

type QoSStats struct {
	BufferUtilization float64
	DroppedPackets    uint64
}

func parseCounters(jsonData []byte) InterfaceCounters {
	// Parse JSON to extract counters
	// Implementation depends on your switch's YANG model
	var counters InterfaceCounters
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err == nil {
		// Extract counter values (implementation depends on YANG model structure)
		// This is a placeholder - adapt to your switch's JSON structure
		if inOctets, ok := data["in-octets"].(float64); ok {
			counters.InOctets = uint64(inOctets)
		}
		if outOctets, ok := data["out-octets"].(float64); ok {
			counters.OutOctets = uint64(outOctets)
		}
	}
	return counters
}

func parseQoSStats(jsonData []byte) QoSStats {
	// Parse QoS statistics
	var qos QoSStats
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err == nil {
		// Extract QoS values (implementation depends on YANG model structure)
		if util, ok := data["buffer-utilization"].(float64); ok {
			qos.BufferUtilization = util
		}
		if drops, ok := data["dropped-packets"].(float64); ok {
			qos.DroppedPackets = uint64(drops)
		}
	}
	return qos
}

// gRPC credentials helper
type loginCreds struct {
	Username string
	Password string
}

func (c *loginCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"username": c.Username,
		"password": c.Password,
	}, nil
}

func (c *loginCreds) RequireTransportSecurity() bool {
	return true
}

func main() {
	configFile := flag.String("config", "/etc/st2110/switches.yaml", "Path to switches configuration")
	listenAddr := flag.String("listen", ":9273", "Prometheus exporter listen address")
	flag.Parse()

	// Load configuration
	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	// Expand password from environment if needed
	for i := range config.Switches {
		if config.Switches[i].Password == "${GNMI_PASSWORD}" {
			config.Switches[i].Password = os.Getenv("GNMI_PASSWORD")
		}
	}

	// Start collectors for each switch
	for _, sw := range config.Switches {
		collector := NewGNMICollector(sw.Target, sw.Username, sw.Password)

		go func(c *GNMICollector, name string) {
			ctx := context.Background()
			log.Printf("Starting gNMI collector for %s (%s)", name, c.target)
			if err := c.Subscribe(ctx); err != nil {
				log.Printf("Subscription error for %s: %v", name, err)
			}
		}(collector, sw.Name)
	}

	// Expose Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting gNMI collector on %s", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
