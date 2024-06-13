package models

type WebsocketDataPayload struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Data any    `json:"data"`
}

var SocketDataPayload = make(chan WebsocketDataPayload, 3)
