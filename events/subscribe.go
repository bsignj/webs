package events

import (
	"strings"
	"webs/ws"
)

type SubscribeIn struct {
	Type string `json:"type"`
}

func SubscribeHandler(hub *ws.Hub, client *ws.Client, event *ws.Event) {
	room := strings.Split(event.Type, ":")[1]
	hub.SubscribeToRoom(room, client)
}
