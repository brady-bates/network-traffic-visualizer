package simulation

import (
	"github.com/paulmach/orb"
	"log"
	"net"
	"networktrafficart/internal/geo"
	_map "networktrafficart/internal/map"
)

type Location struct {
	X, Y     float64
	Lifespan int
}

func NewLocation(ip net.IP, geo geo.GeoService, bounds orb.Bound) Location {
	x, y := getCoordsFromIP(ip, geo, bounds)
	return Location{
		x, y, 10,
	}
}

func getCoordsFromIP(ip net.IP, geo geo.GeoService, bounds orb.Bound) (float64, float64) {
	long, lat, err := geo.LongLatFromIP(ip)
	if err != nil {
		log.Fatal(err)
	}
	x, y := _map.Project(bounds, long, lat)
	return x, y
}
