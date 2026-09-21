package mockdatastream

import (
	"math/rand"
	"net"
	"networktrafficvisualizer/internal/capture"
	"time"
)

func generateRandomIPv4() net.IP {
	o1 := byte(rand.Intn(256))
	o2 := byte(rand.Intn(256))
	o3 := byte(rand.Intn(256))
	o4 := byte(rand.Intn(256))

	return net.IPv4(o1, o2, o3, o4).To4()
}

func Start(events chan capture.PacketData, delayMicros int, batchSize int) {
	micro := time.Duration(delayMicros) * time.Microsecond
	for {
		for range batchSize {
			select {
			case events <- capture.NewPacketData(generateRandomIPv4(), generateRandomIPv4()):
			default:
			}
		}
		time.Sleep(micro)
	}
}
