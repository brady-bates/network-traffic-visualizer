package mockdatastream

import (
	"networktrafficart/internal/capture"
	"networktrafficart/internal/util"
	"time"
)

func Start(events chan capture.PacketData, delayMicros int, batchSize int) {
	micro := time.Duration(delayMicros) * time.Microsecond
	for {
		for range batchSize {
			select {
			case events <- capture.NewPacketData(util.GenerateRandomIPv4(), util.GenerateRandomIPv4()):
			default:
			}
		}
		time.Sleep(micro)
	}
}
