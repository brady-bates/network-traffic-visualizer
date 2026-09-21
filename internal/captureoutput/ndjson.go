package captureoutput

import (
	"context"
	"encoding/json"
	"io"
	"networktrafficvisualizer/internal/capture"
	"time"
)

const SchemaVersion = 1

type PacketEvent struct {
	SchemaVersion      int       `json:"schemaVersion"`
	ObservedAt         time.Time `json:"observedAt"`
	SourceAddress      string    `json:"sourceAddress"`
	DestinationAddress string    `json:"destinationAddress"`
	Direction          string    `json:"direction"`
}

func NewPacketEvent(data capture.Packet, observedAt time.Time) PacketEvent {
	direction := "outbound"
	if data.IsIncoming {
		direction = "inbound"
	}

	return PacketEvent{
		SchemaVersion:      SchemaVersion,
		ObservedAt:         observedAt.UTC(),
		SourceAddress:      data.SrcIP.String(),
		DestinationAddress: data.DstIP.String(),
		Direction:          direction,
	}
}

func StreamNDJSON(ctx context.Context, output io.Writer, packets <-chan capture.Packet, limit int) error {
	encoder := json.NewEncoder(output)
	written := 0

	for {
		select {
		case <-ctx.Done():
			return nil
		case packet, ok := <-packets:
			if !ok {
				return nil
			}
			if err := encoder.Encode(NewPacketEvent(packet, time.Now())); err != nil {
				return err
			}

			written++
			if limit > 0 && written >= limit {
				return nil
			}
		}
	}
}
