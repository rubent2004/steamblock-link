package link

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/coder/websocket"
)

// dispatch enruta un método JSON-RPC al servicio correspondiente. Devuelve el
// resultado (que se serializa como `result`) o un error (que va como `error`).
func (s *Server) dispatch(ctx context.Context, c *websocket.Conn, req rpcRequest) (any, error) {
	switch req.Method {
	case "ports:list":
		return s.serial.ListPorts(s.cfg.SimEnabled)

	case "serial:connect":
		var p struct {
			Path     string `json:"path"`
			BaudRate int    `json:"baudRate"`
		}
		if err := arg(req, 0, &p); err != nil {
			return nil, err
		}
		if !s.claimOwner(c) {
			return nil, errors.New("Otro cliente ya tiene el puerto abierto.")
		}
		h := SerialHandlers{
			OnData:  func(d string) { s.notify(ctx, c, "serial:data", d) },
			OnClose: func() { s.notify(ctx, c, "serial:disconnected", nil) },
		}
		if err := s.serial.Connect(p.Path, p.BaudRate, h); err != nil {
			s.releaseOwner(c)
			return nil, err
		}
		return true, nil

	case "serial:disconnect":
		s.releaseOwner(c)
		if err := s.serial.Disconnect(); err != nil {
			return nil, err
		}
		return true, nil

	case "serial:send":
		var data string
		if err := arg(req, 0, &data); err != nil {
			return nil, err
		}
		return nil, s.serial.Send(data)

	case "upload:sketch":
		var p struct {
			Code  string `json:"code"`
			Board Board  `json:"board"`
			Port  string `json:"port"`
		}
		if err := arg(req, 0, &p); err != nil {
			return nil, err
		}
		return s.compiler.Upload(ctx, p.Code, p.Board, p.Port), nil

	case "compile:check":
		var p struct {
			Code  string `json:"code"`
			Board Board  `json:"board"`
		}
		if err := arg(req, 0, &p); err != nil {
			return nil, err
		}
		return s.compiler.Check(ctx, p.Code, p.Board), nil

	default:
		return nil, errors.New("método no soportado por el Link: " + req.Method)
	}
}

// arg decodifica el parámetro posicional i de la petición en dst.
func arg(req rpcRequest, i int, dst any) error {
	if i >= len(req.Params) {
		return errors.New("faltan parámetros en la petición")
	}
	if err := json.Unmarshal(req.Params[i], dst); err != nil {
		return errors.New("parámetro inválido: " + err.Error())
	}
	return nil
}
