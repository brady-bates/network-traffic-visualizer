package main

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/hajimehoshi/ebiten/v2"
	"log"
	"net"
	"networktrafficart/internal/capture"
	"networktrafficart/internal/capture/mockdatastream"
	"networktrafficart/internal/config"
	"networktrafficart/internal/csv"
	"networktrafficart/internal/display"
	"networktrafficart/internal/geo"
	_map "networktrafficart/internal/map"
	"networktrafficart/internal/simulation"
	"networktrafficart/internal/util"
)

const (
	title             = "Network Traffic Art"
	captureDeviceName = "en0" // TODO add logic to get the best device to capture with
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatal(err)
	}
	conf := config.GetConfig()

	subnet, err := capture.GetInterfaceIPv4SubnetRange(captureDeviceName)
	if err != nil {
		log.Fatal(err)
	}

	handle, err := pcap.OpenLive(captureDeviceName, 65536, true, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	capt := capture.NewCaptureProvider(handle, subnet)

	if conf.EnablePacketCaptureFilter {
		var ipv4 net.IP
		if ipv4, err = capture.GetInterfaceIPv4(captureDeviceName); err != nil {
			log.Fatal(err)
		}

		filter := fmt.Sprintf("%s %s", conf.PacketCaptureFilter, ipv4.String())
		if err = capt.SetHandleBPFFilter(filter); err != nil {
			log.Println("Failed to set packet filter ", err)
		}
	}

	var csvWriterIn chan gopacket.Packet
	if conf.WritePacketsToCSV {
		csvWriterIn = make(chan gopacket.Packet)
		go csv.StreamToCSV(util.GetShutDownCtx(), csvWriterIn, conf.CsvName)
	}

	go capt.StartPacketCapture(csvWriterIn)

	if conf.EnableMockEventStream {
		go mockdatastream.Start(capt.Data, conf.MockEventStreamDelayMicros, conf.MockEventBatchSize)
	}

	geoData := _map.LoadGeoJSON("assets/map/map.geojson")
	geoService := geo.NewGeoService("assets/geolitedb/GeoLite2-City.mmdb")
	sim := simulation.NewSimulation(capt.Data, geoData.Bounds, geoService)
	disp := display.NewDisplay(sim, geoData, geoService)

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle(title)
	ebiten.SetWindowSize(disp.ScreenWidth, disp.ScreenHeight)
	ebiten.SetFullscreen(conf.Fullscreen)

	sim.Init(
		conf.PacketBufferConsumerMaxDelayMicros,
		conf.PacketBufferConsumerAggressionCurve,
	)
	if err = ebiten.RunGameWithOptions(disp, &ebiten.RunGameOptions{
		GraphicsLibrary: ebiten.GraphicsLibraryOpenGL,
	}); err != nil {
		log.Fatal(err)
	}
}
