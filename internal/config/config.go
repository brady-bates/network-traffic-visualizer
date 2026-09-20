package config

import (
	"github.com/joho/godotenv"
	"networktrafficart/internal/util"
	"os"
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
	EnablePacketCaptureFilter           bool
	PacketCaptureFilter                 string
	PacketBufferConsumerAggressionCurve float64
}

// TODO add handling for missing env values?
func LoadConfig() error {
	err := godotenv.Load()
	config = &Config{
		Fullscreen:                          util.IsTrueStr(os.Getenv("FULLSCREEN")),
		EnableMockEventStream:               util.IsTrueStr(os.Getenv("ENABLE_MOCK_EVENT_STREAM")),
		MockEventStreamDelayMicros:          util.ParseToInt(os.Getenv("MOCK_EVENT_STREAM_DELAY_MICROS")),
		MockEventBatchSize:                  util.ParseToInt(os.Getenv("MOCK_EVENT_BATCH_SIZE")),
		PacketBufferConsumerMaxDelayMicros:  util.ParseToInt(os.Getenv("PACKET_BUFFER_CONSUMER_MAX_DELAY_MICROS")),
		WritePacketsToCSV:                   util.IsTrueStr(os.Getenv("WRITE_PACKETS_TO_CSV")),
		CsvName:                             os.Getenv("CSV_NAME"),
		EnablePacketCaptureFilter:           util.IsTrueStr(os.Getenv("ENABLE_PACKET_CAPTURE_FILTER")),
		PacketCaptureFilter:                 strings.TrimSpace(os.Getenv("PACKET_CAPTURE_FILTER")),
		PacketBufferConsumerAggressionCurve: util.ParseToFloat(os.Getenv("PACKET_BUFFER_CONSUMER_AGGRESSION_CURVE")),
	}
	return err
}

func GetConfig() *Config {
	return config
}
