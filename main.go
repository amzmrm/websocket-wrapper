package main

import (
	"net/http"

	"github.com/amzmrm/websocket-wrapper/pkg/ws"
	"github.com/go-chi/chi"
)

// RequestCodes
const (
	Ping            ws.RequestCode = "PING"
	SetMatchingMode ws.RequestCode = "SET_MATCHING_MODE"
	PickUser        ws.RequestCode = "PICK_USER"
)

func main() {
	mux := ws.NewWsRouter()
	mux.HandleFunc(Ping, func(client *ws.Client, param *ws.Request) {
		client.SendResponse(&ws.Response{
			Code:  Ping,
			Error: false,
			Data:  "Pong",
		})
	})
	mux.HandleFunc(SetMatchingMode, SetMatchingModeFunc)
	t := ws.NewTequila()
	t.Use(ws.NewRecovery())
	t.UseWsHandler(mux)

	r := chi.NewRouter()
	r.Get("/conn", ws.NewWebSocketConn)
	http.ListenAndServe(":8080", r)
}

// SetMatchingModeFunc updates user's current matching mode
func SetMatchingModeFunc(client *ws.Client, req *ws.Request) {
	client.SendResponse(&ws.Response{
		Code:  SetMatchingMode,
		Error: false,
		Data:  "Set matching mode success",
	})
}
