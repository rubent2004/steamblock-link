package link

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"sync"

	"github.com/coder/websocket"
)

// Config parametriza el servidor del Link.
type Config struct {
	SimEnabled bool     // expone el puerto serie virtual (STUDIO_SIM_SERIAL)
	Token      string   // si != "", el cliente debe presentarlo en ?token=
	Origins    []string // orígenes extra permitidos además de los built-in
	LibsDir    string   // carpeta de librerías Arduino vendoreadas
	CLIPath    string   // ruta a arduino-cli
}

// Origen de producción siempre permitido (además de localhost).
var builtinOrigins = []string{"https://steamblock.steamroboticsacademy.com"}

var localhostOrigin = regexp.MustCompile(`^https?://(localhost|127\.0\.0\.1)(:\d+)?$`)

// Server enruta las peticiones JSON-RPC del IDE hacia el serie y el compilador.
// El puerto serie tiene dueño único: solo la conexión que lo abrió recibe datos
// y puede cerrarlo (una pestaña del IDE a la vez).
type Server struct {
	cfg      Config
	serial   *Serial
	compiler *Compiler

	mu    sync.Mutex
	owner *websocket.Conn
}

// NewServer construye el servidor con sus servicios (serie + compilador).
func NewServer(cfg Config) *Server {
	return &Server{
		cfg:      cfg,
		serial:   NewSerial(),
		compiler: NewCompiler(cfg.CLIPath, cfg.LibsDir),
	}
}

// Compiler expone el compilador para que main pueda preparar el entorno Arduino.
func (s *Server) Compiler() *Compiler { return s.compiler }

// Handler valida origen/token y actualiza la conexión a WebSocket. Se monta en
// un http.Server que escucha SOLO en loopback (nunca 0.0.0.0).
func (s *Server) Handler(w http.ResponseWriter, r *http.Request) {
	if !s.authorize(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	// Verificamos el origen a mano (arriba), así que saltamos el check del lib.
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	c.SetReadLimit(8 << 20) // 8 MiB: los sketches con imágenes/fuentes pueden pesar.
	s.serve(c)
}

func (s *Server) authorize(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if !s.originAllowed(origin) {
		return false
	}
	if s.cfg.Token != "" && r.URL.Query().Get("token") != s.cfg.Token {
		return false
	}
	return true
}

func (s *Server) originAllowed(origin string) bool {
	if origin == "" || origin == "file://" { // apps empaquetadas / same-origin
		return true
	}
	if localhostOrigin.MatchString(origin) {
		return true
	}
	for _, o := range builtinOrigins {
		if o == origin {
			return true
		}
	}
	for _, o := range s.cfg.Origins {
		if o == origin {
			return true
		}
	}
	return false
}

// serve atiende una conexión hasta que se cierra: lee peticiones, las despacha y
// libera el puerto serie si esta conexión era su dueña.
func (s *Server) serve(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log.Print("[steamblock-link] cliente conectado")

	defer func() {
		if s.releaseOwner(c) {
			_ = s.serial.Disconnect()
		}
		_ = c.CloseNow()
		log.Print("[steamblock-link] cliente desconectado")
	}()

	for {
		_, data, err := c.Read(ctx)
		if err != nil {
			return
		}
		s.handle(ctx, c, data)
	}
}

func (s *Server) handle(ctx context.Context, c *websocket.Conn, data []byte) {
	var req rpcRequest
	if err := json.Unmarshal(data, &req); err != nil || req.JSONRPC != jsonrpcVersion || req.Method == "" {
		s.write(ctx, c, newError(nil, errParse, "JSON-RPC 2.0 inválido"))
		return
	}
	result, err := s.dispatch(ctx, c, req)
	if err != nil {
		s.write(ctx, c, newError(req.ID, errInternal, err.Error()))
		return
	}
	s.write(ctx, c, newResult(req.ID, result))
}

// write serializa y envía un mensaje. coder/websocket permite escrituras
// concurrentes, así que las notificaciones del serie y las respuestas conviven.
func (s *Server) write(ctx context.Context, c *websocket.Conn, msg any) {
	b, err := json.Marshal(msg)
	if err != nil {
		return
	}
	_ = c.Write(ctx, websocket.MessageText, b)
}

func (s *Server) notify(ctx context.Context, c *websocket.Conn, method string, params any) {
	s.write(ctx, c, rpcNotification{JSONRPC: jsonrpcVersion, Method: method, Params: params})
}

// --- Propiedad del puerto serie (dueño único) ---

func (s *Server) claimOwner(c *websocket.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.owner != nil && s.owner != c {
		return false
	}
	s.owner = c
	return true
}

func (s *Server) releaseOwner(c *websocket.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.owner == c {
		s.owner = nil
		return true
	}
	return false
}
