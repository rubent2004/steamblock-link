package link

import (
	"sync"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

// SerialHandlers son los callbacks que el servidor conecta a un puerto abierto:
// onData por cada trozo recibido, onClose cuando el puerto se cierra o falla.
type SerialHandlers struct {
	OnData  func(string)
	OnClose func()
}

// Serial encapsula el puerto activo (uno a la vez, como el TS original). Toda la
// mutación pasa por el mutex porque connect/disconnect/send y la goroutine de
// lectura corren en hilos distintos.
type Serial struct {
	mu     sync.Mutex
	port   serial.Port
	sim    *simDevice // no-nil cuando el puerto simulado está activo
	simGen int        // generación: invalida lecturas de una sesión ya cerrada
}

// NewSerial crea el gestor de puerto serie.
func NewSerial() *Serial { return &Serial{} }

// ListPorts enumera los puertos reales; si simEnabled, antepone el virtual.
func (s *Serial) ListPorts(simEnabled bool) ([]SerialPortInfo, error) {
	out := []SerialPortInfo{}
	if simEnabled {
		out = append(out, simPortInfo)
	}
	// GetDetailedPortsList da fabricante/producto; si falla, caemos a la lista simple.
	details, err := enumerator.GetDetailedPortsList()
	if err == nil {
		for _, p := range details {
			out = append(out, SerialPortInfo{Path: p.Name, Manufacturer: p.Product, Label: p.Name})
		}
		return out, nil
	}
	names, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		out = append(out, SerialPortInfo{Path: name, Label: name})
	}
	return out, nil
}

// Connect abre opts.Path a opts.BaudRate (o el simulador si es el puerto SIM),
// cerrando antes cualquier puerto previo. Los datos entrantes van a h.OnData.
func (s *Serial) Connect(path string, baud int, h SerialHandlers) error {
	if err := s.Disconnect(); err != nil {
		return err
	}
	if isSimPort(path) {
		s.mu.Lock()
		s.simGen++
		s.sim = startSim(h)
		s.mu.Unlock()
		return nil
	}

	port, err := serial.Open(path, &serial.Mode{BaudRate: baud})
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.port = port
	s.simGen++
	gen := s.simGen
	s.mu.Unlock()

	// Lectura en segundo plano: Read se desbloquea con error al cerrar el puerto,
	// lo que termina la goroutine y dispara OnClose una sola vez por sesión.
	go s.readLoop(port, gen, h)
	return nil
}

func (s *Serial) readLoop(port serial.Port, gen int, h SerialHandlers) {
	buf := make([]byte, 4096)
	for {
		n, err := port.Read(buf)
		if !s.sessionAlive(gen) {
			return // desconexión provocada por nosotros: OnClose ya no aplica.
		}
		if err != nil || n == 0 {
			s.mu.Lock()
			if s.simGen == gen {
				s.port = nil
				s.simGen++
			}
			s.mu.Unlock()
			if h.OnClose != nil {
				h.OnClose()
			}
			return
		}
		if h.OnData != nil {
			h.OnData(string(buf[:n]))
		}
	}
}

// Disconnect cierra el puerto (real o simulado) si hay alguno abierto.
func (s *Serial) Disconnect() error {
	s.mu.Lock()
	port, sim := s.port, s.sim
	s.port, s.sim = nil, nil
	s.simGen++ // invalida la sesión: readLoop/sim se callan sin disparar OnClose.
	s.mu.Unlock()

	if sim != nil {
		sim.stop()
	}
	if port != nil {
		return port.Close()
	}
	return nil
}

// Send escribe una línea (con salto) al puerto activo. Sin puerto, es no-op.
func (s *Serial) Send(data string) error {
	s.mu.Lock()
	port, sim := s.port, s.sim
	s.mu.Unlock()
	if sim != nil {
		sim.send(data)
		return nil
	}
	if port == nil {
		return nil
	}
	_, err := port.Write([]byte(data + "\n"))
	return err
}

func (s *Serial) sessionAlive(gen int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.simGen == gen
}
