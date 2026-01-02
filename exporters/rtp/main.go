package main

import (
	"flag"
	"log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"st2110-exporter/rtp"
	"st2110-exporter/exporter"
)

type Config struct {
	Streams []rtp.StreamConfig `yaml:"streams"`
}

func main() {
	configFile := flag.String("config", "config/streams.yaml", "Path to streams configuration")
	listenAddr := flag.String("listen", ":9100", "Prometheus exporter listen address")
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

	// Create exporter
	exp := exporter.NewST2110Exporter()

	// Add streams
	for _, streamConfig := range config.Streams {
		if err := exp.AddStream(streamConfig); err != nil {
			log.Printf("Failed to add stream %s: %v", streamConfig.StreamID, err)
			continue
		}
		log.Printf("Added stream: %s (%s)", streamConfig.Name, streamConfig.Multicast)
	}

	// Start HTTP server
	log.Fatal(exp.ServeHTTP(*listenAddr))
}

