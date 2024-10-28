package ws

import (
	"webs/config"
	log "webs/pkg/logger"

	"github.com/alphadose/haxmap"
	"go.uber.org/zap"
)

// Room represents a chat room where clients can join and receive messages.
type Room struct {
	name       string
	clients    *haxmap.Map[string, *Client]
	register   chan *Client
	unregister chan *Client
	send       chan *Event
	workers    int
}

// newRoom creates a new Room instance with the specified name and configuration.
func newRoom(name string, cfg *config.Room) *Room {
	room := &Room{
		name:       name,
		clients:    haxmap.New[string, *Client](uintptr(cfg.BufferedClientSize)),
		register:   make(chan *Client, cfg.BufferedRegisterSize),
		unregister: make(chan *Client, cfg.BufferedUnregisterSize),
		send:       make(chan *Event, cfg.BufferedMessageSize),
		workers:    cfg.BufferedWorkersSize,
	}

	// Start the worker goroutines to handle broadcasted messages.
	for i := 0; i < room.workers; i++ {
		go room.consumeBroadcastedMessage(i)
	}
	// Start the main room management routine.
	go room.run()

	return room
}

// run is the main loop that manages client registration and unregistration in the room.
func (room *Room) run() {
	for {
		select {
		// Add a client to the room.
		case client := <-room.register:
			room.clients.Set(client.ID, client)
			client.rooms.Set(room.name, room)
			log.ZLogger.Debug("[MSG-Room] Client entered the room", zap.String("ClientID", client.ID), zap.String("Room", room.name))

		// Remove a client from the room.
		case client := <-room.unregister:
			if _, ok := room.clients.Get(client.ID); ok {
				room.clients.Del(client.ID)
				client.rooms.Del(room.name)
				log.ZLogger.Debug("[MSG-Room] Client left the room", zap.String("ClientID", client.ID), zap.String("Room", room.name))
			}
		}
	}
}

// consumeBroadcastedMessage is a worker that sends broadcasted events to all clients in the room.
func (room *Room) consumeBroadcastedMessage(workerID int) {
	for event := range room.send {
		room.clients.ForEach(func(clientID string, client *Client) bool {
			select {
			// Send the event to the client.
			case client.send <- event:
			// Log a warning if the event could not be sent.
			default:
				log.ZLogger.Warn("[MSG-Room] Event lost for client", zap.Int("WorkerID", workerID), zap.String("ClientID", client.ID))
			}
			return true
		})
	}
}
