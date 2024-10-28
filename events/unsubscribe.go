package events

import (
	"strings"
	"webs/ws"
)

type UnsubscribeIn struct {
	Type string `json:"type"`
}

func UnsubscribeHandler(hub *ws.Hub, client *ws.Client, event *ws.Event) {
	room := strings.Split(event.Type, ":")[1]
	hub.UnsubscribeFromRoom(room, client)
}
