package arista

import (
	"context"
	"fmt"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/prometheus/client_golang/prometheus"
)

// Note: This requires the base GNMICollector from exporters/gnmi
// In a real implementation, you would import it or embed it properly

type AristaEOSCollector struct {
	// Embedded GNMICollector (would be imported from parent package)
	// *GNMICollector

	target string

	// Arista-specific metrics
	hwQueueDrops    *prometheus.CounterVec
	ptpLockStatus   *prometheus.GaugeVec
	igmpGroups      *prometheus.GaugeVec
	tcamUtilization *prometheus.GaugeVec
}

func NewAristaEOSCollector(target, username, password string) *AristaEOSCollector {
	return &AristaEOSCollector{
		target: target,

		hwQueueDrops: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "arista_hw_queue_drops_total",
				Help: "Hardware queue drops (critical for ST 2110)",
			},
			[]string{"switch", "interface", "queue"},
		),

		ptpLockStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "arista_ptp_lock_status",
				Help: "PTP lock status (1=locked, 0=unlocked)",
			},
			[]string{"switch", "domain"},
		),

		igmpGroups: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "arista_igmp_snooping_groups",
				Help: "IGMP snooping multicast groups per VLAN",
			},
			[]string{"switch", "vlan"},
		),

		tcamUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "arista_tcam_utilization_percent",
				Help: "TCAM utilization (multicast routing table)",
			},
			[]string{"switch", "table"},
		),
	}
}

// Subscribe to Arista-specific paths
func (c *AristaEOSCollector) SubscribeArista(ctx context.Context, client gnmi.GNMIClient) error {
	// Arista EOS uses vendor-specific YANG models
	subscribeReq := &gnmi.SubscribeRequest{
		Request: &gnmi.SubscribeRequest_Subscribe{
			Subscribe: &gnmi.SubscriptionList{
				Mode: gnmi.SubscriptionList_STREAM,
				Subscription: []*gnmi.Subscription{
					// Hardware queue drops (Arista-specific path)
					{
						Path: &gnmi.Path{
							Origin: "arista", // Arista vendor origin
							Elem: []*gnmi.PathElem{
								{Name: "eos"},
								{Name: "arista-exp-eos-qos"},
								{Name: "qos"},
								{Name: "interfaces"},
								{Name: "interface", Key: map[string]string{"name": "*"}},
								{Name: "queues"},
								{Name: "queue", Key: map[string]string{"queue-id": "*"}},
								{Name: "state"},
								{Name: "dropped-pkts"},
							},
						},
						Mode:           gnmi.SubscriptionMode_SAMPLE,
						SampleInterval: 1000000000, // 1 second
					},
					// PTP status (if using Arista as PTP Boundary Clock)
					{
						Path: &gnmi.Path{
							Origin: "arista",
							Elem: []*gnmi.PathElem{
								{Name: "eos"},
								{Name: "arista-exp-eos-ptp"},
								{Name: "ptp"},
								{Name: "instances"},
								{Name: "instance", Key: map[string]string{"instance-id": "default"}},
								{Name: "state"},
							},
						},
						Mode: gnmi.SubscriptionMode_ON_CHANGE,
					},
					// IGMP snooping state
					{
						Path: &gnmi.Path{
							Origin: "arista",
							Elem: []*gnmi.PathElem{
								{Name: "eos"},
								{Name: "arista-exp-eos-igmpsnooping"},
								{Name: "igmp-snooping"},
								{Name: "vlans"},
								{Name: "vlan", Key: map[string]string{"vlan-id": "*"}},
								{Name: "state"},
							},
						},
						Mode: gnmi.SubscriptionMode_ON_CHANGE,
					},
				},
				Encoding: gnmi.Encoding_JSON_IETF,
			},
		},
	}

	// Start subscription
	stream, err := client.Subscribe(ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	if err := stream.Send(subscribeReq); err != nil {
		return fmt.Errorf("failed to send subscription: %w", err)
	}

	// Process updates
	for {
		response, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("stream error: %w", err)
		}

		c.handleAristaUpdate(response)
	}
}

func (c *AristaEOSCollector) handleAristaUpdate(response *gnmi.SubscribeResponse) {
	switch resp := response.Response.(type) {
	case *gnmi.SubscribeResponse_Update:
		notification := resp.Update

		for _, update := range notification.Update {
			path := update.Path
			value := update.Val

			// Parse Arista-specific hardware queue drops
			if path.Origin == "arista" && len(path.Elem) > 7 {
				if path.Elem[7].Name == "dropped-pkts" {
					ifaceName := path.Elem[4].Key["name"]
					queueID := path.Elem[6].Key["queue-id"]

					drops := value.GetUintVal()
					c.hwQueueDrops.WithLabelValues(c.target, ifaceName, queueID).Add(float64(drops))

					// Alert if drops detected (should be ZERO for ST 2110!)
					if drops > 0 {
						fmt.Printf("⚠️  Hardware queue drops on %s interface %s queue %s: %d packets\n",
							c.target, ifaceName, queueID, drops)
					}
				}
			}
		}
	}
}
