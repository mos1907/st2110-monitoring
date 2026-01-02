package grassvalley

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

type KFrameExporter struct {
	baseURL string // http://kframe-ip
	apiKey  string

	// K-Frame specific metrics
	cardStatus         *prometheus.GaugeVec
	cardTemperature    *prometheus.GaugeVec
	videoInputStatus   *prometheus.GaugeVec
	audioChannelStatus *prometheus.GaugeVec
	crosspointStatus   *prometheus.GaugeVec
}

func NewKFrameExporter(baseURL, apiKey string) *KFrameExporter {
	return &KFrameExporter{
		baseURL: baseURL,
		apiKey:  apiKey,

		cardStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "grassvalley_kframe_card_status",
				Help: "K-Frame card status (1=OK, 0=fault)",
			},
			[]string{"chassis", "slot", "card_type"},
		),

		cardTemperature: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "grassvalley_kframe_card_temperature_celsius",
				Help: "K-Frame card temperature",
			},
			[]string{"chassis", "slot", "card_type"},
		),

		videoInputStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "grassvalley_kframe_video_input_status",
				Help: "Video input signal status (1=present, 0=no signal)",
			},
			[]string{"chassis", "slot", "input"},
		),

		audioChannelStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "grassvalley_kframe_audio_channel_status",
				Help: "Audio channel status (1=present, 0=silent)",
			},
			[]string{"chassis", "slot", "channel"},
		),

		crosspointStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "grassvalley_kframe_crosspoint_count",
				Help: "Number of active crosspoints (router connections)",
			},
			[]string{"chassis", "router_level"},
		),
	}
}

// K-Frame REST API endpoints
func (e *KFrameExporter) Collect() error {
	// Get chassis inventory
	chassis, err := e.getChassis()
	if err != nil {
		return err
	}

	for _, ch := range chassis {
		// Get card status for each slot
		cards, err := e.getCards(ch.ID)
		if err != nil {
			continue
		}

		for _, card := range cards {
			// Update card status
			e.cardStatus.WithLabelValues(ch.Name, card.Slot, card.Type).Set(boolToFloat(card.Healthy))
			e.cardTemperature.WithLabelValues(ch.Name, card.Slot, card.Type).Set(card.Temperature)

			// Get video input status (for ST 2110 receivers)
			if card.Type == "IPDENSITY" || card.Type == "IPG-3901" {
				inputs, err := e.getVideoInputs(ch.ID, card.Slot)
				if err != nil {
					continue
				}

				for _, input := range inputs {
					e.videoInputStatus.WithLabelValues(
						ch.Name, card.Slot, input.Name,
					).Set(boolToFloat(input.SignalPresent))
				}
			}
		}

		// Get router crosspoint count
		crosspoints, err := e.getCrosspoints(ch.ID)
		if err != nil {
			continue
		}

		e.crosspointStatus.WithLabelValues(ch.Name, "video").Set(float64(crosspoints.VideoCount))
		e.crosspointStatus.WithLabelValues(ch.Name, "audio").Set(float64(crosspoints.AudioCount))
	}

	return nil
}

// K-Frame REST API client methods
func (e *KFrameExporter) makeRequest(endpoint string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v2/%s", e.baseURL, endpoint)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-API-Key", e.apiKey)
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

func (e *KFrameExporter) getChassis() ([]Chassis, error) {
	data, err := e.makeRequest("chassis")
	if err != nil {
		return nil, err
	}

	var result struct {
		Chassis []Chassis `json:"chassis"`
	}
	json.Unmarshal(data, &result)

	return result.Chassis, nil
}

func (e *KFrameExporter) getCards(chassisID string) ([]Card, error) {
	data, err := e.makeRequest(fmt.Sprintf("chassis/%s/cards", chassisID))
	if err != nil {
		return nil, err
	}

	var result struct {
		Cards []Card `json:"cards"`
	}
	json.Unmarshal(data, &result)

	return result.Cards, nil
}

func (e *KFrameExporter) getVideoInputs(chassisID, slot string) ([]VideoInput, error) {
	endpoint := fmt.Sprintf("chassis/%s/cards/%s/inputs", chassisID, slot)
	data, err := e.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var result struct {
		Inputs []VideoInput `json:"inputs"`
	}
	json.Unmarshal(data, &result)

	return result.Inputs, nil
}

func (e *KFrameExporter) getCrosspoints(chassisID string) (*Crosspoints, error) {
	data, err := e.makeRequest(fmt.Sprintf("chassis/%s/crosspoints", chassisID))
	if err != nil {
		return nil, err
	}

	var crosspoints Crosspoints
	json.Unmarshal(data, &crosspoints)

	return &crosspoints, nil
}

type Chassis struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Card struct {
	Slot        string  `json:"slot"`
	Type        string  `json:"type"`
	Healthy     bool    `json:"healthy"`
	Temperature float64 `json:"temperature"`
}

type VideoInput struct {
	Name          string `json:"name"`
	SignalPresent bool   `json:"signal_present"`
	Format        string `json:"format"`
}

type Crosspoints struct {
	VideoCount int `json:"video_count"`
	AudioCount int `json:"audio_count"`
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func main() {
	baseURL := flag.String("url", "http://kframe-ip", "K-Frame base URL")
	apiKey := flag.String("apikey", "", "K-Frame API key")
	listenAddr := flag.String("listen", ":9400", "Prometheus exporter listen address")
	flag.Parse()

	exporter := NewKFrameExporter(*baseURL, *apiKey)

	// Register metrics
	prometheus.MustRegister(exporter.cardStatus)
	prometheus.MustRegister(exporter.cardTemperature)
	prometheus.MustRegister(exporter.videoInputStatus)
	prometheus.MustRegister(exporter.audioChannelStatus)
	prometheus.MustRegister(exporter.crosspointStatus)

	// Collect metrics periodically
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for range ticker.C {
			if err := exporter.Collect(); err != nil {
				log.Printf("Collection error: %v", err)
			}
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting K-Frame exporter on %s", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
