package simulation

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/paulmach/orb"
	"math"
	"networktrafficvisualizer/internal/capture"
	"networktrafficvisualizer/internal/geo"
	"slices"
	"sync"
	"time"
)

type Simulation struct {
	CapturePackets    chan capture.Packet
	Locations         []*Location
	mut               sync.RWMutex
	OffScreenDistance float32
	locationBuffer    chan Location
	GeoService        geo.GeoService
	MapBounds         orb.Bound
}

func NewSimulation(cd chan capture.Packet, bounds orb.Bound, geo geo.GeoService) *Simulation {
	return &Simulation{
		CapturePackets:    cd,
		Locations:         []*Location{},
		mut:               sync.RWMutex{},
		OffScreenDistance: 25,
		locationBuffer:    make(chan Location, 50000),
		GeoService:        geo,
		MapBounds:         bounds,
	}
}

func (s *Simulation) Init(PacketBufferConsumerMaxDelayMicros int, PacketBufferConsumerAggressionCurve float64) {
	go s.WatchEventChannel()
	go s.CreateLocationsFromBuffer(
		PacketBufferConsumerAggressionCurve,
		PacketBufferConsumerMaxDelayMicros,
	)
}

func (s *Simulation) Tick() {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.tickLocations()
}

func (s *Simulation) tickLocations() {
	s.Locations = slices.DeleteFunc(s.Locations, func(loc *Location) bool {
		loc.Lifespan -= 1
		return loc.Lifespan <= 0
	})
}

func (s *Simulation) DrawLocations(screen *ebiten.Image, circle *ebiten.Image) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	opts := &ebiten.DrawImageOptions{}
	for _, p := range s.Locations {
		opts.GeoM.Reset()
		opts.GeoM.Translate(float64(p.X), float64(p.Y))
		screen.DrawImage(circle, opts)
	}
}

func (s *Simulation) AddToLocations(l *Location) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.Locations = append(s.Locations, l)
}

func (s *Simulation) WatchEventChannel() {
	for packet := range s.CapturePackets {
		locs := []Location{
			NewLocation(packet.SrcIP, s.GeoService, s.MapBounds),
			NewLocation(packet.DstIP, s.GeoService, s.MapBounds),
		}

		for _, loc := range locs {
			if s.containsLocation(loc) {
				continue
			}
			select {
			case s.locationBuffer <- loc:
			default:
			}
		}
	}
}

func (s *Simulation) containsLocation(target Location) bool {
	s.mut.RLock()
	defer s.mut.RUnlock()

	return slices.ContainsFunc(s.Locations, func(l *Location) bool {
		return l.X == target.X && l.Y == target.Y
	})
}

func clampValue(val, min, max float64) float64 {
	return math.Max(min, math.Min(val, max))
}

func (s *Simulation) CreateLocationsFromBuffer(aggressionCurve float64, maxWatcherDelay int) {
	curve := clampValue(aggressionCurve, 0.0, math.Inf(+1))
	capacity := float64(cap(s.locationBuffer))
	minDelay := 0.0
	maxDelay := float64(maxWatcherDelay)

	var location Location
	for location = range s.locationBuffer {
		count := float64(len(s.locationBuffer))
		fullness := count / (capacity * .6)
		modulationFactor := math.Pow(fullness, curve)
		modulatedDelay := maxDelay + modulationFactor*(minDelay-maxDelay)
		micro := time.Duration(modulatedDelay) * time.Microsecond

		s.AddToLocations(&location)

		time.Sleep(micro)
	}
}

// TODO find a new visual representation of a packet
// Packet/trail, leaves a faded trail
// Arc of some kind, brief fade in and out - dot traveling along arc?
