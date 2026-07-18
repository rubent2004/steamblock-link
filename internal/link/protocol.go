// Package link implementa el companion local de STEAMBLOCK Studio: un servidor
// WebSocket que habla JSON-RPC 2.0 con el IDE web y delega en el puerto serie y
// en arduino-cli para monitorear, compilar y subir sketches a placas reales.
//
// Es el port a Go del antiguo steamblock-link (TypeScript/Node): mismo contrato
// (shared/ipc.ts + shared/jsonrpc.ts), pero como binario estático único, sin
// runtime de Node ni ventana de consola. Ver README.md.
package link

import "encoding/json"

// Versión de JSON-RPC que hablamos. El contrato es idéntico al del cliente web.
const jsonrpcVersion = "2.0"

// Códigos de error JSON-RPC 2.0 estándar (mismos que shared/jsonrpc.ts).
const (
	errParse    = -32700
	errInternal = -32603
)

// rpcRequest es una petición entrante del IDE. params es posicional (un array),
// se decodifica por método en el dispatcher.
type rpcRequest struct {
	JSONRPC string            `json:"jsonrpc"`
	ID      json.RawMessage   `json:"id"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
}

// rpcResponse es la respuesta a una petición (éxito o error, nunca ambos).
type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

// rpcNotification es un evento push del servidor al cliente (sin id).
type rpcNotification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func newResult(id json.RawMessage, result any) rpcResponse {
	return rpcResponse{JSONRPC: jsonrpcVersion, ID: id, Result: result}
}

func newError(id json.RawMessage, code int, msg string) rpcResponse {
	return rpcResponse{JSONRPC: jsonrpcVersion, ID: id, Error: &rpcError{Code: code, Message: msg}}
}

// --- Tipos del dominio (espejo de shared/ipc.ts y shared/types/board.ts) ---

// SerialPortInfo describe un puerto serie para el selector del IDE.
type SerialPortInfo struct {
	Path         string `json:"path"`
	Manufacturer string `json:"manufacturer"`
	Label        string `json:"label"`
}

// UploadResult es el resultado de compilar/subir (ok + mensaje para el usuario).
type UploadResult struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// Board es la placa seleccionada en el IDE. Solo necesitamos el FQBN para
// compilar/subir; el resto de campos del contrato se ignoran sin romper el JSON.
type Board struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	FQBN string `json:"fqbn"`
}
