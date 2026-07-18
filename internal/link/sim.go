package link

import (
	"fmt"
	"math/rand"
	"time"
)

// Dispositivo serie virtual para desarrollar y demostrar sin placa física.
// Se activa con STUDIO_SIM_SERIAL=1 y aparece en la lista como un puerto más.
var simPortInfo = SerialPortInfo{
	Path:         "SIM://steamblock",
	Manufacturer: "STEAMBLOCK",
	Label:        "Simulador STEAMBLOCK (virtual)",
}

func isSimPort(path string) bool { return path == simPortInfo.Path }

// simDevice emite telemetría falsa (temperatura/distancia) una vez por segundo y
// hace eco de lo que se le envía, hasta que se llama a stop().
type simDevice struct {
	done chan struct{}
	h    SerialHandlers
}

func startSim(h SerialHandlers) *simDevice {
	d := &simDevice{done: make(chan struct{}), h: h}
	if h.OnData != nil {
		h.OnData("SIM: dispositivo virtual STEAMBLOCK conectado\n")
	}
	go d.loop()
	return d
}

func (d *simDevice) loop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	start := time.Now()
	for {
		select {
		case <-d.done:
			return
		case now := <-ticker.C:
			if d.h.OnData == nil {
				continue
			}
			secs := int(now.Sub(start).Seconds())
			temp := 20 + rand.Float64()*8
			dist := rand.Intn(200)
			d.h.OnData(fmt.Sprintf("t=%ds temp=%.1fC dist=%dcm\n", secs, temp, dist))
		}
	}
}

func (d *simDevice) send(data string) {
	if d.h.OnData != nil {
		d.h.OnData("echo: " + data + "\n")
	}
}

func (d *simDevice) stop() {
	select {
	case <-d.done: // ya cerrado
	default:
		close(d.done)
	}
}
