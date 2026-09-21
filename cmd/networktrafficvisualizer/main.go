package main

import (
	"context"
	"fmt"
	"github.com/google/gopacket"
	"github.com/hajimehoshi/ebiten/v2"
	"log"
	"networktrafficvisualizer/internal/capture"
	"networktrafficvisualizer/internal/capture/mockdatastream"
	"networktrafficvisualizer/internal/config"
	"networktrafficvisualizer/internal/csv"
	"networktrafficvisualizer/internal/display"
	"networktrafficvisualizer/internal/geo"
	pkgmap "networktrafficvisualizer/internal/map"
	"networktrafficvisualizer/internal/simulation"
	"runtime"
)

const (
	title = "Network Traffic Art"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if _, err := config.LoadConfig(); err != nil {
		log.Fatal(err)
	}
	conf := config.GetConfig()

	device, err := capture.FindDevice(conf.CaptureInterface)
	if err != nil {
		log.Fatal(err)
	}

	handle, err := capture.OpenDevice(device)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	log.Printf("Capturing on %s with IPv4 subnet %s", device.DisplayName(), device.Subnet)
	capt := capture.NewCaptureProvider(handle, device.Subnet)

	if conf.EnablePacketCaptureFilter {
		filter := fmt.Sprintf("%s %s", conf.PacketCaptureFilter, device.IPv4.String())
		if err = capt.SetHandleBPFFilter(filter); err != nil {
			log.Println("Failed to set packet filter ", err)
		}
	}

	var csvWriterIn chan gopacket.Packet
	if conf.WritePacketsToCSV {
		csvWriterIn = make(chan gopacket.Packet)
		go csv.StreamToCSV(ctx, csvWriterIn, conf.CsvName)
	}

	go capt.StartPacketCapture(csvWriterIn)

	if conf.EnableMockEventStream {
		go mockdatastream.Start(capt.Packets, conf.MockEventStreamDelayMicros, conf.MockEventBatchSize)
	}

	geoData := pkgmap.LoadGeoJSON("assets/map/map.geojson")
	geoService := geo.NewGeoService("assets/geolitedb/GeoLite2-City.mmdb")
	sim := simulation.NewSimulation(capt.Packets, geoData.Bounds, geoService)
	disp := display.NewDisplay(sim, geoData, geoService, cancel)

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle(title)
	ebiten.SetWindowSize(disp.ScreenWidth, disp.ScreenHeight)
	ebiten.SetFullscreen(conf.Fullscreen)

	sim.Init(
		conf.PacketBufferConsumerMaxDelayMicros,
		conf.PacketBufferConsumerAggressionCurve,
	)
	runOptions := &ebiten.RunGameOptions{}
	if runtime.GOOS == "darwin" {
		runOptions.GraphicsLibrary = ebiten.GraphicsLibraryOpenGL
	}
	if err = ebiten.RunGameWithOptions(disp, runOptions); err != nil {
		log.Fatal(err)
	}
}
