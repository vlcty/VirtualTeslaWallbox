package main

import (
	"encoding/json"
	// "fmt"
	"github.com/vlcty/TeslaWallbox"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	GRID_FREQUENCY_MIN float64 = 49.2
	GRID_FREQUENCY_MAX float64 = 51.8

	GRID_VOLTAGE_MIN float64 = 227.0
	GRID_VOLTAGE_MAX float64 = 230.5
)

type VirtualWallbox struct {
	vitals        *teslaWallbox.Vitals
	lifetimeStats *teslaWallbox.LifetimeStats
	version       *teslaWallbox.Version

	mutex *sync.Mutex

	voltageChannel   chan float64
	frequencyChannel chan float64
}

func InitVirtualWallbox() *VirtualWallbox {
	vw := &VirtualWallbox{
		vitals:        &teslaWallbox.Vitals{},
		lifetimeStats: &teslaWallbox.LifetimeStats{},
		version:       &teslaWallbox.Version{},
		mutex:         &sync.Mutex{}}

	vw.vitals.ContactorClosed = false
	vw.vitals.VehicleConnected = false

	vw.version.FirmwareVersion = "virtual wallbox v1.0"
	vw.version.PartNumber = "virtual wallbox"
	vw.version.SerialNumber = "1234567890"

	vw.startUptimeTicker()
	vw.startGridGenerator()

	return vw
}

func (vw *VirtualWallbox) startUptimeTicker() {
	ticker := time.NewTicker(time.Second)

	go func() {
		for {
			<-ticker.C

			vw.vitals.Uptime++
		}
	}()
}

func newRandomFloatWithBoundaries(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func (vw *VirtualWallbox) startGridGenerator() {
	vw.vitals.GridFrequency = newRandomFloatWithBoundaries(GRID_FREQUENCY_MIN, GRID_FREQUENCY_MAX)
	vw.vitals.GridVoltage = newRandomFloatWithBoundaries(GRID_VOLTAGE_MIN, GRID_VOLTAGE_MAX)

	ticker := time.NewTicker(time.Second * 5)

	go func() {
		for {
			<-ticker.C

			vw.mutex.Lock()
			vw.vitals.GridFrequency = newRandomFloatWithBoundaries(GRID_FREQUENCY_MIN, GRID_FREQUENCY_MAX)
			vw.vitals.GridVoltage = newRandomFloatWithBoundaries(GRID_VOLTAGE_MIN, GRID_VOLTAGE_MAX)
			vw.mutex.Unlock()
		}
	}()
}

func (vw *VirtualWallbox) processAPIRequest(writer http.ResponseWriter, request *http.Request, thing interface{}) {
	vw.mutex.Lock()
	defer vw.mutex.Unlock()
	defer request.Body.Close()

	json.NewEncoder(writer).Encode(thing)
}

func (vw *VirtualWallbox) processVitalsRequest(writer http.ResponseWriter, request *http.Request) {
	vw.processAPIRequest(writer, request, vw.vitals)
}

func (vw *VirtualWallbox) processLifetimeRequest(writer http.ResponseWriter, request *http.Request) {
	vw.processAPIRequest(writer, request, vw.lifetimeStats)
}

func (vw *VirtualWallbox) processVersionRequest(writer http.ResponseWriter, request *http.Request) {
	vw.processAPIRequest(writer, request, vw.version)
}

func main() {
	virtualWallbox := InitVirtualWallbox()

	http.HandleFunc("/api/1/vitals", virtualWallbox.processVitalsRequest)
	http.HandleFunc("/api/1/lifetime", virtualWallbox.processLifetimeRequest)
	http.HandleFunc("/api/1/version", virtualWallbox.processVersionRequest)

	http.ListenAndServe(":8080", nil)
}
