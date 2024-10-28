package ws

import (
	"net/http"
	"webs/config"
	log "webs/pkg/logger"

	"github.com/alphadose/haxmap"
	"go.uber.org/zap"

	"github.com/gorilla/websocket"
)

// The websocketUpgrader is used to upgrade the HTTP connection to a WebSocket connection.
var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Commented for testing purposes
	//CheckOrigin: func(r *http.Request) bool {
	//	origin := r.Header.Get("Origin")
	//	switch origin {
	//	case "http://localhost:8383":
	//		return true
	//	default:
	//		return false
	//	}
	//},
}

// Hub represents a WebSocket hub.
type Hub struct {
	clients    *haxmap.Map[string, *Client]      // clients contains all connected clients.
	rooms      *haxmap.Map[string, *Room]        // rooms contains all created rooms.
	events     *haxmap.Map[string, EventHandler] // events contains all registered event handlers.
	register   chan *Client                      // register is used to register new clients.
	unregister chan *Client                      // unregister is used to unregister clients.
	send       chan *Event                       // send is used to send events to all clients.
}

// NewHub creates a new Hub instance.
func NewHub(cfg *config.Hub) *Hub {
	hub := &Hub{
		clients:    haxmap.New[string, *Client](uintptr(cfg.BufferedClientSize)),
		rooms:      haxmap.New[string, *Room](uintptr(cfg.BufferedRoomSize)),
		events:     haxmap.New[string, EventHandler](uintptr(cfg.BufferedEventSize)),
		register:   make(chan *Client, cfg.BufferedRegisterSize),
		unregister: make(chan *Client, cfg.BufferedUnregisterSize),
		send:       make(chan *Event, cfg.BufferedMessageSize),
	}

	// Start the hub's run goroutine.
	go hub.run()
	return hub
}

// run is the main loop of the hub.
func (hub *Hub) run() {
	for {
		select {
		// Register a new client.
		case client := <-hub.register:
			hub.clients.Set(client.ID, client)
			log.ZLogger.Debug("[MSG-HUB] Client registriert", zap.String("ClientID", client.ID))

		// Unregister a client.
		case client := <-hub.unregister:
			if _, ok := hub.clients.Get(client.ID); ok {
				// Disconnect from all rooms.
				client.rooms.ForEach(func(key string, room *Room) bool {
					room.clients.Del(client.ID)
					client.rooms.Del(room.name)
					log.ZLogger.Debug("[MSG-Room] Client verlässt den Raum", zap.String("ClientID", client.ID), zap.String("Room", room.name))
					return true
				})

				// Disconnect from the hub.
				close(client.send)
				client.conn.Close()
				hub.clients.Del(client.ID)
				log.ZLogger.Debug("[MSG-HUB] Client abgemeldet", zap.String("ClientID", client.ID))
			}

		// Broadcast a message to all clients.
		case event := <-hub.send:
			hub.clients.ForEach(func(clientID string, client *Client) bool {
				client.send <- event
				return true
			})
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// OnConnect upgrades the HTTP connection to a WebSocket connection and initializes a new client.
func (hub *Hub) OnConnect(w http.ResponseWriter, r *http.Request, cfg *config.Client) (*Client, error) {
	// Upgrade the regular connection to a WebSocket connection.
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.ZLogger.Error("[ERR-HUB] Akzeptieren der Verbindung:", zap.Error(err))
		return nil, err
	}

	// Create a new client and add it to the hub.
	client := newClient(hub, conn, cfg)
	hub.register <- client

	// Start the client's read and write goroutines.
	go client.readMessage()
	go client.writeMessage()

	return client, nil
}

// On registers an event handler for a specific event.
func (hub *Hub) On(event string, eventHandler EventHandler) {
	hub.events.Set(event, eventHandler)
}

// CreateRooms creates multiple rooms.
func (hub *Hub) CreateRooms(roomNames []string, cfg *config.Room) {
	for _, name := range roomNames {
		room := newRoom(name, cfg)
		hub.rooms.Set(name, room)
	}
}

// //////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// BroadcastToAll broadcasts an event to all clients.
func (hub *Hub) BroadcastToAll(event *Event) {
	hub.send <- event
}

// BroadcastToRoom broadcasts an event to all clients in the specified room.
func (hub *Hub) BroadcastToRoom(roomName string, event *Event) {
	if room, ok := hub.rooms.Get(roomName); ok {
		room.send <- event
	}
}

// SendToClient sends an event to a specific client.
func (hub *Hub) SendToClient(client *Client, event *Event) {
	client.send <- event
}

// SubscribeToRoom subscribes a client to a room.
func (hub *Hub) SubscribeToRoom(roomName string, client *Client) {
	if room, ok := hub.rooms.Get(roomName); ok {
		room.register <- client
	}
}

// UnsubscribeFromRoom unsubscribes a client from a room.
func (hub *Hub) UnsubscribeFromRoom(roomName string, client *Client) {
	if room, ok := hub.rooms.Get(roomName); ok {
		room.unregister <- client
	}
}
