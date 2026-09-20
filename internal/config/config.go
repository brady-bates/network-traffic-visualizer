package config

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"strings"
)

var (
	config *Config
)

type Config struct {
	Fullscreen                          bool
	EnableMockEventStream               bool
	MockEventStreamDelayMicros          int
	MockEventBatchSize                  int
	PacketBufferConsumerMaxDelayMicros  int
	WritePacketsToCSV                   bool
	CsvName                             string
	CaptureInterface                    string
	EnablePacketCaptureFilter           bool
	PacketCaptureFilter                 string
	PacketBufferConsumerAggressionCurve float64
}

func isTrueStr(s string) bool {
	return strings.TrimSpace(s) == "true"
}

func parseToInt(s string) (int, error) {
	if strings.TrimSpace(s) == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}

func parseToFloat(s string) (float64, error) {
	if strings.TrimSpace(s) == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

func LoadConfig() error {
	err := godotenv.Load()

	mockDelay, errInt1 := parseToInt(os.Getenv("MOCK_EVENT_STREAM_DELAY_MICROS"))
	if errInt1 != nil {
		return errInt1
	}

	batchSize, errInt2 := parseToInt(os.Getenv("MOCK_EVENT_BATCH_SIZE"))
	if errInt2 != nil {
		return errInt2
	}

	maxDelay, errInt3 := parseToInt(os.Getenv("PACKET_BUFFER_CONSUMER_MAX_DELAY_MICROS"))
	if errInt3 != nil {
		return errInt3
	}

	curve, errFloat := parseToFloat(os.Getenv("PACKET_BUFFER_CONSUMER_AGGRESSION_CURVE"))
	if errFloat != nil {
		return errFloat
	}

	config = &Config{
		Fullscreen:                          isTrueStr(os.Getenv("FULLSCREEN")),
		EnableMockEventStream:               isTrueStr(os.Getenv("ENABLE_MOCK_EVENT_STREAM")),
		MockEventStreamDelayMicros:          mockDelay,
		MockEventBatchSize:                  batchSize,
		PacketBufferConsumerMaxDelayMicros:  maxDelay,
		WritePacketsToCSV:                   isTrueStr(os.Getenv("WRITE_PACKETS_TO_CSV")),
		CsvName:                             os.Getenv("CSV_NAME"),
		CaptureInterface:                    strings.TrimSpace(os.Getenv("CAPTURE_INTERFACE")),
		EnablePacketCaptureFilter:           isTrueStr(os.Getenv("ENABLE_PACKET_CAPTURE_FILTER")),
		PacketCaptureFilter:                 strings.TrimSpace(os.Getenv("PACKET_CAPTURE_FILTER")),
		PacketBufferConsumerAggressionCurve: curve,
	}
	return err
}

func GetConfig() *Config {
	return config
}
