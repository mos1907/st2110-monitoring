// Synthetic Test Stream Generator - Generates test ST 2110 streams for monitoring validation
package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type TestStreamGenerator struct {
	multicast    string
	port         int
	format       string
	bitrate      uint64
	injectErrors bool
	errorRate    float64

	conn      *net.UDPConn
	seqNumber uint16
	timestamp uint32
	ssrc      uint32
}

func NewTestStreamGenerator(multicast string, port int, format string) *TestStreamGenerator {
	return &TestStreamGenerator{
		multicast: multicast,
		port:      port,
		format:    format,
		bitrate:   2200000000, // 2.2Gbps for 1080p60
		ssrc:      rand.Uint32(),
	}
}

// Generate synthetic ST 2110 stream for testing
func (g *TestStreamGenerator) Start() error {
	// Resolve multicast address
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", g.multicast, g.port))
	if err != nil {
		return fmt.Errorf("failed to resolve address: %v", err)
	}

	// Create UDP connection
	g.conn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("failed to create UDP connection: %v", err)
	}

	log.Printf("Generating test stream to %s:%d (format: %s)", g.multicast, g.port, g.format)

	// Calculate packet rate for format
	// 1080p60: ~90,000 packets/second
	packetRate := 90000
	if g.format == "720p60" {
		packetRate = 45000
	} else if g.format == "1080p50" {
		packetRate = 75000
	}

	interval := time.Second / time.Duration(packetRate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	for {
		select {
		case <-sigChan:
			log.Println("Received interrupt, stopping generator...")
			return nil
		case <-ticker.C:
			if err := g.sendPacket(); err != nil {
				log.Printf("Error sending packet: %v", err)
			}
		}
	}
}

func (g *TestStreamGenerator) sendPacket() error {
	// Inject errors if enabled
	if g.injectErrors && rand.Float64()*100 < g.errorRate {
		// Skip packet (simulate loss)
		g.seqNumber++
		return nil
	}

	// Build RTP packet
	rtpLayer := &layers.RTP{
		Version:        2,
		Padding:        false,
		Extension:      false,
		Marker:         false,
		PayloadType:    96, // Dynamic
		SequenceNumber: g.seqNumber,
		Timestamp:      g.timestamp,
		SSRC:           g.ssrc,
	}

	// Generate dummy payload (1400 bytes typical for ST 2110)
	payload := make([]byte, 1400)
	if _, err := rand.Read(payload); err != nil {
		return err
	}

	// Serialize packet
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: false, // UDP checksum optional
	}

	if err := gopacket.SerializeLayers(buf, opts, rtpLayer, gopacket.Payload(payload)); err != nil {
		return err
	}

	// Send
	if _, err := g.conn.Write(buf.Bytes()); err != nil {
		return err
	}

	// Increment counters
	g.seqNumber++
	g.timestamp += 1500 // 90kHz / 60fps = 1500

	return nil
}

// Enable error injection (for testing packet loss detection)
func (g *TestStreamGenerator) InjectErrors(rate float64) {
	g.injectErrors = true
	g.errorRate = rate
	log.Printf("Injecting %.3f%% packet loss", rate)
}

// Stop generating
func (g *TestStreamGenerator) Stop() {
	if g.conn != nil {
		g.conn.Close()
	}
}

func main() {
	multicast := flag.String("multicast", "239.1.1.100", "Multicast address")
	port := flag.Int("port", 20000, "UDP port")
	format := flag.String("format", "1080p60", "Video format (1080p60, 720p60, 1080p50)")
	errorRate := flag.Float64("error-rate", 0, "Packet loss rate percentage (0-100)")
	flag.Parse()

	// Allow override from environment
	if envMulticast := os.Getenv("MULTICAST"); envMulticast != "" {
		multicast = &envMulticast
	}
	if envPort := os.Getenv("PORT"); envPort != "" {
		var p int
		if n, err := fmt.Sscanf(envPort, "%d", &p); err == nil && n == 1 {
			*port = p
		}
	}
	if envFormat := os.Getenv("FORMAT"); envFormat != "" {
		format = &envFormat
	}

	generator := NewTestStreamGenerator(*multicast, *port, *format)

	if *errorRate > 0 {
		generator.InjectErrors(*errorRate)
	}

	if err := generator.Start(); err != nil {
		log.Fatalf("Failed to start generator: %v", err)
	}

	generator.Stop()
}
