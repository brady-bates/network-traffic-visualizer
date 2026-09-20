package display

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
	"networktrafficart/internal/geo"
	"networktrafficart/internal/lifecycle"
	_map "networktrafficart/internal/map"
	"networktrafficart/internal/simulation"
)

const (
	screenWidth, screenHeight = 1920, 1080
)

type Display struct {
	simulation      *simulation.Simulation
	ScreenWidth     int
	ScreenHeight    int
	baseCircleImage *ebiten.Image
	screenBuffer    *ebiten.Image
	geoJsonData     _map.MapData
	geoService      geo.GeoService
	mapProjection   *ebiten.Image
}

func NewDisplay(s *simulation.Simulation, geoData _map.MapData, geoService geo.GeoService) *Display {
	circleImage := ebiten.NewImage(6, 6)
	vector.FillCircle(circleImage, 3, 3, 3, color.White, true)

	return &Display{
		simulation:      s,
		ScreenWidth:     screenWidth,
		ScreenHeight:    screenHeight,
		baseCircleImage: circleImage,
		screenBuffer:    ebiten.NewImage(screenWidth, screenHeight),
		geoJsonData:     geoData,
		geoService:      geoService,
		mapProjection:   nil,
	}
}

func (d *Display) Update() error {
	//fmt.Printf("fps: %f tps: %f\n", ebiten.ActualFPS(), ebiten.ActualTPS())

	if ebiten.IsWindowBeingClosed() {
		ebiten.SetWindowClosingHandled(true)
		lifecycle.GetShutDownCtx().Cancel()

		return ebiten.Termination
	}

	d.simulation.Tick()
	return nil
}

func (d *Display) Draw(screen *ebiten.Image) {
	if d.mapProjection == nil {
		d.mapProjection = _map.DrawMap(d.geoJsonData, ebiten.NewImage(d.ScreenWidth, d.ScreenHeight))
	}

	d.screenBuffer.Fill(_map.OceanColor)
	d.screenBuffer.DrawImage(d.mapProjection, nil)

	d.simulation.DrawLocations(d.screenBuffer, d.baseCircleImage)

	screen.DrawImage(d.screenBuffer, nil)
}

func (d *Display) Layout(w, h int) (int, int) {
	return d.ScreenWidth, d.ScreenHeight
}
