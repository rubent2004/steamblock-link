package link

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// dialTestServer levanta el handler en un httptest y devuelve un cliente WS.
func dialTestServer(t *testing.T, cfg Config) (*websocket.Conn, func()) {
	t.Helper()
	srv := NewServer(cfg)
	ts := httptest.NewServer(http.HandlerFunc(srv.Handler))
	url := "ws" + strings.TrimPrefix(ts.URL, "http")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		ts.Close()
		t.Fatalf("dial: %v", err)
	}
	return c, func() { c.CloseNow(); ts.Close() }
}

func rawRead(t *testing.T, c *websocket.Conn) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, data, err := c.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return data
}

// isResponse distingue una respuesta (lleva `id`) de una notificación (no).
func isResponse(b []byte) bool {
	var m struct {
		ID *json.RawMessage `json:"id"`
	}
	return json.Unmarshal(b, &m) == nil && m.ID != nil
}

func write(t *testing.T, c *websocket.Conn, id int, method string, params ...any) {
	t.Helper()
	b, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Write(ctx, websocket.MessageText, b); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// call envía una petición y devuelve su respuesta, saltando las notificaciones
// que puedan intercalarse antes (p.ej. el primer serial:data del simulador).
func call(t *testing.T, c *websocket.Conn, id int, method string, params ...any) rpcResponse {
	t.Helper()
	write(t, c, id, method, params...)
	for range 10 {
		b := rawRead(t, c)
		if isResponse(b) {
			var r rpcResponse
			if err := json.Unmarshal(b, &r); err != nil {
				t.Fatalf("unmarshal respuesta %s: %v", b, err)
			}
			return r
		}
	}
	t.Fatalf("no llegó respuesta para %q", method)
	return rpcResponse{}
}

func TestPortsListIncludesSimulator(t *testing.T) {
	c, done := dialTestServer(t, Config{SimEnabled: true})
	defer done()

	resp := call(t, c, 1, "ports:list")
	if resp.Error != nil {
		t.Fatalf("error inesperado: %v", resp.Error)
	}
	ports, _ := json.Marshal(resp.Result)
	if !strings.Contains(string(ports), simPortInfo.Path) {
		t.Fatalf("la lista no incluye el puerto simulado: %s", ports)
	}
}

func TestSimSerialPushesData(t *testing.T) {
	c, done := dialTestServer(t, Config{SimEnabled: true})
	defer done()

	write(t, c, 1, "serial:connect", map[string]any{"path": simPortInfo.Path, "baudRate": 9600})

	// Esperamos ver tanto la respuesta (result=true) como al menos una
	// notificación serial:data del dispositivo virtual, en cualquier orden.
	sawResp, sawData := false, false
	for i := 0; i < 10 && (!sawResp || !sawData); i++ {
		b := rawRead(t, c)
		if isResponse(b) {
			var r rpcResponse
			json.Unmarshal(b, &r)
			if r.Error != nil || r.Result != true {
				t.Fatalf("connect falló: result=%v err=%v", r.Result, r.Error)
			}
			sawResp = true
			continue
		}
		var n rpcNotification
		json.Unmarshal(b, &n)
		if n.Method == "serial:data" {
			if s, ok := n.Params.(string); ok && s != "" {
				sawData = true
			}
		}
	}
	if !sawResp || !sawData {
		t.Fatalf("faltó respuesta o datos: sawResp=%v sawData=%v", sawResp, sawData)
	}
}

func TestSecondClientCannotClaimSerial(t *testing.T) {
	cfg := Config{SimEnabled: true}
	srv := NewServer(cfg)
	ts := httptest.NewServer(http.HandlerFunc(srv.Handler))
	defer ts.Close()
	url := "ws" + strings.TrimPrefix(ts.URL, "http")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c1, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("dial c1: %v", err)
	}
	defer c1.CloseNow()
	c2, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("dial c2: %v", err)
	}
	defer c2.CloseNow()

	if r := call(t, c1, 1, "serial:connect", map[string]any{"path": simPortInfo.Path, "baudRate": 9600}); r.Result != true {
		t.Fatalf("c1 debería adueñarse del puerto: %v / %v", r.Result, r.Error)
	}
	r := call(t, c2, 2, "serial:connect", map[string]any{"path": simPortInfo.Path, "baudRate": 9600})
	if r.Error == nil {
		t.Fatalf("c2 no debería poder abrir el puerto ya tomado")
	}
}
