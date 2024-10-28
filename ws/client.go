package ws

import (
	"bytes"
	"time"
	"webs/config"
	log "webs/pkg/logger"

	"github.com/alphadose/haxmap"
	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	pongWait       = 60 * time.Second    // pongWait is the time to wait before sending a ping message.
	pingInterval   = (pongWait * 9) / 10 // pingInterval is the time between sending ping messages.
	maxMessageSize = 512                 // maxMessageSize is the maximum size of a message.
)

var (
	newline = []byte{'\n'} // newline is the newline character.
	space   = []byte{' '}  // space is the space character.
)

// Client represents a WebSocket client.
type Client struct {
	ID    string                     // ID is a unique identifier for the client.
	hub   *Hub                       // hub is the Hub that the client is connected to.
	conn  *websocket.Conn            // conn is the WebSocket connection to the client.
	rooms *haxmap.Map[string, *Room] // rooms is a map of rooms that the client is joined to.
	send  chan *Event                // send is a channel for sending messages to the client.
}

// newClient creates a new client.
func newClient(hub *Hub, conn *websocket.Conn, cfg *config.Client) *Client {
	return &Client{
		ID:    uuid.New().String(),
		hub:   hub,
		conn:  conn,
		rooms: haxmap.New[string, *Room](uintptr(cfg.BufferedRoomSize)),
		send:  make(chan *Event, cfg.BufferedMessageSize),
	}
}

// readMessage reads messages from the client and broadcasts them to the hub.
func (client *Client) readMessage() {
	defer func() {
		client.hub.unregister <- client
	}()

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		// Read a message from the client.
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.ZLogger.Info("[ERR-Client] Nachricht lesen", zap.Error(err))
			}
			return
		}

		// Trim the message.
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		// Parse the message into an Event.
		event, err := NewEventFromRaw(message)
		if err != nil {
			log.ZLogger.Error("[ERR-Client] Event erstellen", zap.Error(err))
			continue
		}
		log.ZLogger.Debug("[MSG-Client] Eingehende Nachricht", zap.String("Message", string(message)))

		// Handle the event.
		if eventHandler, ok := client.hub.events.Get(event.Type); ok {
			eventHandler(event)
		}
	}
}

// writeMessage writes messages to the client.
func (client *Client) writeMessage() {
	// Create a ticker for sending pings.
	ticker := time.NewTicker(pingInterval)

	defer func() {
		ticker.Stop()
		client.hub.unregister <- client
	}()

	for {
		select {
		// Send a ping to the client.
		case <-ticker.C:
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.ZLogger.Info("[ERR-Client] Fehler Ping senden", zap.Error(err))
				return
			}

		// Write a message to the client.
		case event, ok := <-client.send:
			if !ok {
				if err := client.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.ZLogger.Info("[ERR-Client] Websocket Verbindung geschlossen", zap.Error(err))
				}
				return
			}

			message, err := event.Raw()
			if err != nil {
				log.ZLogger.Error("[ERR-Client] Nachricht erstellen", zap.Error(err))
				continue
			}
			log.ZLogger.Debug("[MSG-Client] Ausgehende Nachricht", zap.String("Message", string(message)))

			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.ZLogger.Warn("[ERR-Client] Fehler beim Schreiben der Nachricht", zap.Error(err))
				return
			}
		}
	}
}
