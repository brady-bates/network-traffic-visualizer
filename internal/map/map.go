package _map

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"image/color"
	"log"
	"math"
	"os"
)

var (
	landmassColor = color.RGBA{R: 25, G: 40, B: 60, A: 255} // dark slate blue
	borderColor   = color.RGBA{R: 40, G: 60, B: 90, A: 255} // slightly lighter, very subtle
	OceanColor    = color.RGBA{R: 8, G: 15, B: 35, A: 255}  // near black navy
)

type MapData struct {
	Features []*geojson.Feature
	Bounds   orb.Bound
}

func DrawMap(data MapData, screenBuffer *ebiten.Image) *ebiten.Image {
	for _, f := range data.Features {
		switch geom := f.Geometry.(type) {
		case orb.Polygon:
			drawPolygon(data.Bounds, screenBuffer, geom)
		case orb.MultiPolygon:
			for _, poly := range geom {
				drawPolygon(data.Bounds, screenBuffer, poly)
			}
		}
	}
	return screenBuffer
}

// Mercator projection
func Project(bounds orb.Bound, lon, lat float64) (float64, float64) {
	minLon := bounds.Min.Lon()
	minLat := bounds.Min.Lat()
	maxLon := bounds.Max.Lon()
	maxLat := bounds.Max.Lat()

	mercMin := translateToMercator(minLat)
	mercMax := translateToMercator(maxLat)
	mercN := translateToMercator(lat)

	padding := 40.0

	w, h := ebiten.WindowSize()
	width := float64(w) - padding*2
	height := float64(h) - padding*2

	x := padding + (lon-minLon)/(maxLon-minLon)*width
	y := padding + (mercMax-mercN)/(mercMax-mercMin)*height

	return x, y
}

func translateToMercator(latDeg float64) float64 {
	r := latDeg * math.Pi / 180
	return math.Log(math.Tan(math.Pi/4 + r/2))
}

func drawPolygon(bounds orb.Bound, screen *ebiten.Image, poly orb.Polygon) {
	for _, ring := range poly {
		// skip undrawable polygons
		if len(ring) < 2 {
			continue
		}

		var path vector.Path

		x0, y0 := Project(bounds, ring[0].Lon(), ring[0].Lat())
		path.MoveTo(float32(x0), float32(y0))

		for _, pt := range ring[1:] {
			x, y := Project(bounds, pt.Lon(), pt.Lat())
			path.LineTo(float32(x), float32(y))
		}
		path.Close()

		fillOpts := &vector.DrawPathOptions{
			AntiAlias: true,
		}
		fillOpts.ColorScale.ScaleWithColor(landmassColor)
		vector.FillPath(screen, &path, &vector.FillOptions{}, fillOpts)

		borderOpts := &vector.DrawPathOptions{
			AntiAlias: true,
		}
		borderOpts.ColorScale.ScaleWithColor(borderColor)
		vector.StrokePath(screen, &path, &vector.StrokeOptions{
			Width: 1.0,
		}, borderOpts)
	}
}

func LoadGeoJSON(path string) MapData {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	fc, err := geojson.UnmarshalFeatureCollection(data)
	if err != nil {
		log.Fatal(err)
	}

	bound := orb.Bound{
		Min: orb.Point{180, 90},
		Max: orb.Point{-180, -90},
	}

	for _, f := range fc.Features {
		if f.Geometry == nil {
			continue
		}

		// Skip geometries below -60 Lat
		if f.Geometry.Bound().Max.Lat() < -60 {
			continue
		}

		switch geom := f.Geometry.(type) {
		case orb.Polygon:
			expandBounds(&bound, geom.Bound())
		case orb.MultiPolygon:
			expandBounds(&bound, geom.Bound())
		}
	}

	if bound.Min[1] < -85.05112 {
		bound.Min[1] = -85.05112
	}
	if bound.Max[1] > 85.05112 {
		bound.Max[1] = 85.05112
	}

	return MapData{fc.Features, bound}
}

func expandBounds(bound *orb.Bound, expandTo orb.Bound) {
	if expandTo.Min.Lon() < bound.Min.Lon() {
		bound.Min[0] = expandTo.Min.Lon()
	}
	if expandTo.Min.Lat() < bound.Min.Lat() {
		bound.Min[1] = expandTo.Min.Lat()
	}
	if expandTo.Max.Lon() > bound.Max.Lon() {
		bound.Max[0] = expandTo.Max.Lon()
	}
	if expandTo.Max.Lat() > bound.Max.Lat() {
		bound.Max[1] = expandTo.Max.Lat()
	}
}
